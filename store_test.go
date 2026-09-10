//go:build linux || darwin || windows

package main

import (
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoreRestartDedupAndAtomicCheckpoint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "log.sqlite")
	s, err := openEventStore(path)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Now().Truncate(time.Millisecond)
	e := radarEvent{Kind: "TOKENS", ID: testID, At: at.Add(-time.Hour), Observed: at, Output: 800, Reasoning: 516}
	p := sourcePosition{Path: "rollout.jsonl", Offset: 400, PrefixLength: 128, PrefixHash: "hash"}
	if ok, err := s.append(&e, "line1", "run1", &p); err != nil || !ok {
		t.Fatalf("append: %v %v", ok, err)
	}
	first := e.Sequence
	if err := s.db.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = openEventStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.db.Close()
	p.Offset = 800
	if ok, err := s.append(&e, "line1", "run2", &p); err != nil || ok {
		t.Fatalf("duplicate: %v %v", ok, err)
	}
	saved, ok, err := s.position(testID)
	if err != nil || !ok || saved.Offset != 400 {
		t.Fatalf("duplicate changed checkpoint: %+v %v", saved, err)
	}
	// An event must roll back if its checkpoint cannot commit.
	_, err = s.db.Exec(`CREATE TRIGGER fail_checkpoint BEFORE UPDATE ON sources BEGIN SELECT RAISE(ABORT,'test checkpoint failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := s.append(&e, "line2", "run2", &p); err == nil || ok {
		t.Fatal("checkpoint failure accepted")
	}
	history, err := s.historySince(at.Add(-time.Second), at)
	if err != nil || len(history) != 1 || history[0].event.Sequence != first || !history[0].event.At.Equal(e.At) {
		t.Fatalf("rollback/source time: %+v %v", history, err)
	}
	_, err = s.db.Exec(`DROP TRIGGER fail_checkpoint`)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := s.append(&e, "line2", "run2", &p); err != nil || !ok {
		t.Fatalf("same values, new event: %v %v", ok, err)
	}
	saved, ok, err = s.position(testID)
	if err != nil || !ok || saved.Offset != 800 {
		t.Fatalf("checkpoint: %+v %v", saved, err)
	}
}

func TestPeak30ExactColorsExpiryAndIndependentRange(t *testing.T) {
	for _, tc := range []struct {
		value uint64
		color string
	}{{0, "yellow"}, {515, "yellow"}, {516, "red"}, {999, "yellow"}, {1000, "green"}, {1001, "yellow"}, {1034, "yellow"}, {1552, "yellow"}} {
		t.Run(tc.color+"_"+time.Duration(tc.value).String(), func(t *testing.T) {
			d := newDashboard()
			defer d.Close()
			now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
			d.now = func() time.Time { return now }
			if s := d.snapshotRange(15); s.Peak30 != nil || s.PeakColor != "none" {
				t.Fatal("empty peak")
			}
			d.record("TOKENS", testID, "peak", now, tokenCount{output: 2000, reasoning: tc.value})
			now = now.Add(30 * time.Minute)
			for _, minutes := range []int{15, 60, 1440, 10080} {
				s := d.snapshotRange(minutes)
				if s.Peak30 == nil || s.Peak30.Reasoning != tc.value || s.PeakColor != tc.color {
					t.Fatalf("range %d: %+v", minutes, s)
				}
				if minutes == 15 && s.Responses != 0 {
					t.Fatal("15m totals include 30m peak")
				}
			}
			now = now.Add(time.Millisecond)
			if s := d.snapshotRange(10080); s.Peak30 != nil || s.PeakColor != "none" || s.Responses != 1 {
				t.Fatal("expired peak/history incorrect")
			}
		})
	}
	d := newDashboard()
	defer d.Close()
	now := time.Now().Truncate(time.Millisecond)
	d.now = func() time.Time { return now }
	d.record("TOKENS", testID, "maximum", now, tokenCount{reasoning: 1000})
	for i := 0; i < 250; i++ {
		d.record("TOKENS", testID, "latest", now, tokenCount{reasoning: 516})
	}
	s := d.snapshotRange(15)
	if s.Peak30.Reasoning != 1000 || s.PeakColor != "green" || len(s.Events) != 200 {
		t.Fatal("peak incorrectly follows latest/bounded feed")
	}
}

func TestHistoryRanges516AndPagination(t *testing.T) {
	d := newDashboard()
	defer d.Close()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	observed := now
	d.now = func() time.Time { return observed }
	for _, age := range []time.Duration{8 * 24 * time.Hour, 7 * 24 * time.Hour, 24 * time.Hour, time.Hour, 15 * time.Minute, 0} {
		observed = now.Add(-age)
		for _, value := range []uint64{516, 1000, 1034, 1552} {
			d.record("TOKENS", testID, "history", observed, tokenCount{output: 2000, reasoning: value})
		}
	}
	observed = now
	for _, tc := range []struct {
		minutes int
		n       uint64
	}{{15, 8}, {60, 12}, {1440, 16}, {10080, 20}} {
		s := d.snapshotRange(tc.minutes)
		if s.Responses != tc.n || s.Count516 != tc.n/4 || s.Output != tc.n*2000 || s.Median != 1017 || s.P95 != 1552 {
			t.Fatalf("range %d incorrect: %+v", tc.minutes, s)
		}
		var n, c uint64
		for _, b := range s.Series {
			n += b.Responses
			c += b.Count516
		}
		if n != s.Responses || c != s.Count516 {
			t.Fatal("bucket totals differ")
		}
	}
	for i := 0; i < 250; i++ {
		d.record("TOKENS", secondID, "516", now, tokenCount{reasoning: 516})
	}
	cursor := uint64(0)
	seen := map[uint64]bool{}
	for {
		p, err := d.store.eventPage(now.Add(-24*time.Hour), now, cursor, true, secondID)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range p.Events {
			if e.Reasoning != 516 || e.ID != secondID || seen[e.Sequence] {
				t.Fatal("wrong or duplicate page item")
			}
			seen[e.Sequence] = true
		}
		if p.Next == 0 {
			break
		}
		cursor = p.Next
	}
	if len(seen) != 250 {
		t.Fatalf("paged %d events", len(seen))
	}
}

func TestMonitorRestartResumesCompleteLines(t *testing.T) {
	f := fixture(t)
	f.metadata(testID, "resume")
	f.lock(testID)
	// A known file starts at EOF on first attachment.
	f.append(testID, "{}\n")
	path := filepath.Join(f.home, "data", "log.sqlite")
	start := func() (*monitor, *dashboard) {
		d, err := newPersistentDashboard(path, io.Discard)
		f.must(err)
		m, err := newMonitor(f.home, &printer{out: &f.out}, &f.err)
		f.must(err)
		m.dashboard = d
		m.readBuffer = make([]byte, 17)
		m.queueActivation(testID, activationAttach, time.Now())
		return m, d
	}
	m, d := start()
	line := tokenLine(800, 516)
	half := len(line) / 2
	f.append(testID, line+line[:half])
	m.readAvailable(m.sessions[testID])
	p, ok := d.resumePosition(testID)
	if !ok || p.Offset != int64(3+len(line)) || d.snapshot().Responses != 1 {
		t.Fatalf("partial checkpoint: %+v", p)
	}
	m.Close()
	f.must(d.Close())
	// Complete a partial response and add an identical-count response while offline.
	f.append(testID, line[half:]+line)
	m, d = start()
	defer m.Close()
	defer d.Close()
	s := d.snapshot()
	if s.Responses != 3 || s.Count516 != 3 || s.Reasoning != 1548 {
		t.Fatalf("restart catchup: %+v", s)
	}
	p, ok = d.resumePosition(testID)
	if !ok || p.Offset != int64(3+3*len(line)) {
		t.Fatalf("resumed checkpoint: %+v", p)
	}
	m.readAvailable(m.sessions[testID])
	if d.snapshot().Responses != 3 {
		t.Fatal("repeated read duplicates")
	}
	// Replaying a source does not double-count its existing lines.
	m.sessions[testID].offset = 0
	m.readAvailable(m.sessions[testID])
	if d.snapshot().Responses != 3 {
		t.Fatal("replayed source duplicates")
	}
	if strings.Contains(f.err.text(), "ERROR") {
		t.Fatal(f.err.text())
	}
}

func TestCoverageDoesNotFillDowntime(t *testing.T) {
	d := newDashboard()
	defer d.Close()
	base := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	now := base
	d.now = func() time.Time { return now }
	d.record("MONITOR_START", "", "", now, tokenCount{})
	now = base.Add(30 * time.Second)
	d.record("HEARTBEAT", "", "", now, tokenCount{})
	now = base.Add(time.Minute)
	d.record("MONITOR_STOP", "", "", now, tokenCount{})
	d.runID = "second-run"
	now = base.Add(10 * time.Minute)
	d.record("MONITOR_START", "", "", now, tokenCount{})
	now = now.Add(20 * time.Second)
	s := d.snapshotRange(15)
	if s.Covered != 80 {
		t.Fatalf("downtime counted: %f seconds", s.Covered)
	}
	var covered float64
	for _, b := range s.Series {
		covered += b.Covered
	}
	if covered != s.Covered {
		t.Fatal("bucket coverage mismatch")
	}
}

func TestHistoryDatabaseFailureIsVisible(t *testing.T) {
	d := newDashboard()
	defer d.Close()
	_, err := d.store.db.Exec(`CREATE TRIGGER reject_events BEFORE INSERT ON events BEGIN SELECT RAISE(ABORT,'test storage failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	d.record("TOKENS", testID, "not committed", time.Now(), tokenCount{reasoning: 516})
	s := d.snapshotRange(1440)
	if s.Storage != "error" || s.Responses != 0 || s.Peak30 != nil {
		t.Fatalf("failed write misrepresented: %+v", s)
	}
}
