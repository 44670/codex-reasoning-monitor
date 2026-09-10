//go:build linux || darwin || windows

package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDashboardLifecycleAndCatchUp(t *testing.T) {
	f := fixture(t)
	p := &printer{out: &f.out}
	m, err := newMonitor(f.home, p, &f.err)
	f.must(err)
	defer m.Close()
	m.dashboard = newDashboard()
	defer m.dashboard.Close()
	lock := f.lock(testID)
	at := time.Now()
	m.queueActivation(testID, activationStart, at)
	snap := m.dashboard.snapshot()
	if len(snap.Sessions) != 1 || snap.Sessions[0].Status != "waiting" {
		t.Fatalf("missing waiting session: %+v", snap.Sessions)
	}
	f.metadata(testID, "Radar test")
	f.append(testID, tokenLine(123, 45))
	m.retryPending(at.Add(time.Second))
	snap = m.dashboard.snapshot()
	if snap.Output != 123 || snap.Reasoning != 45 || snap.Responses != 1 || snap.Sessions[0].Status != "active" {
		t.Fatalf("activation/catch-up not reflected: %+v", snap)
	}
	last := snap.Events[len(snap.Events)-1]
	if last.Kind != "TOKENS" || last.At.Equal(last.Observed) || last.Title != "Radar test" {
		t.Fatalf("source timestamp/title lost: %+v", last)
	}
	var output, reasoning uint64
	for _, b := range snap.Series {
		output += b.Output
		reasoning += b.Reasoning
	}
	if output != 123 || reasoning != 45 {
		t.Fatal("arrival-time chart lost catch-up event")
	}
	f.must(lock.Close())
	m.deactivate(testID, time.Now())
	if got := m.dashboard.snapshot().Sessions[0].Status; got != "stopped" {
		t.Fatalf("status = %s", got)
	}
}

func TestDashboardBoundedHistoryAndConcurrentReaders(t *testing.T) {
	d := newDashboard()
	defer d.Close()
	d.record("START", testID, "first", time.Now(), tokenCount{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 300; j++ {
				_ = d.snapshot()
			}
		}()
	}
	for i := 0; i < 300; i++ {
		d.record("TOKENS", testID, "renamed", time.Now(), tokenCount{output: 10, reasoning: 4})
	}
	wg.Wait()
	s := d.snapshot()
	if len(s.Events) != 200 || len(s.Series) != 61 || len(s.Sessions[0].History) != 24 || s.Output != 3000 || s.Reasoning != 1200 {
		t.Fatal("history caps or cumulative totals are incorrect")
	}
	s.Sessions[0].History[0] = 999
	if d.snapshot().Sessions[0].History[0] != 4 {
		t.Fatal("snapshot shares mutable history")
	}
	for i := 0; i < 205; i++ {
		d.record("STOP", fmt.Sprint(i), "old", time.Now(), tokenCount{})
	}
	s = d.snapshot()
	if len(s.Sessions) != 101 {
		t.Fatalf("retained %d sessions, want 100 stopped + 1 active", len(s.Sessions))
	}
	if s.Peak30 == nil || s.Peak30.Reasoning != 4 || s.Peak30.Output != 10 {
		t.Fatal("latest token response lost after event feed rollover")
	}
	s.Peak30.Reasoning = 999
	if d.snapshot().Peak30.Reasoning != 4 {
		t.Fatal("snapshot shares mutable latest token response")
	}
}

func TestDashboardHTTP(t *testing.T) {
	d := newDashboard()
	defer d.Close()
	d.record("START", testID, "<script>alert('title')</script>", time.Now(), tokenCount{})
	for _, tc := range []struct {
		path, host, origin, method string
		status                     int
	}{
		{"/", dashboardAddress, "", "GET", 200},
		{"/app.js", dashboardAddress, "", "GET", 200},
		{"/style.css", dashboardAddress, "", "GET", 200},
		{"/api/snapshot", dashboardAddress, "", "GET", 200},
		{"/api/snapshot", "localhost:5927", "http://localhost:5927", "GET", 200},
		{"/api/snapshot", "evil.example:5927", "", "GET", 403},
		{"/api/snapshot", dashboardAddress, "https://evil.example", "GET", 403},
		{"/api/snapshot", dashboardAddress, "", "POST", 405},
		{"/missing", dashboardAddress, "", "GET", 404},
	} {
		t.Run(tc.path+tc.host+tc.origin+tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://"+tc.host+tc.path, nil)
			if tc.origin != "" {
				r.Header.Set("Origin", tc.origin)
			}
			w := httptest.NewRecorder()
			d.handler().ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			if tc.status == 200 && tc.path == "/api/snapshot" {
				var s radarSnapshot
				if err := json.Unmarshal(w.Body.Bytes(), &s); err != nil {
					t.Fatal(err)
				}
				if len(s.Sessions) != 1 || strings.Contains(w.Body.String(), "<script>") {
					t.Fatal("bad/unsafe snapshot")
				}
				if w.Header().Get("Access-Control-Allow-Origin") != "" {
					t.Fatal("local data exposed via CORS")
				}
			}
		})
	}
}
