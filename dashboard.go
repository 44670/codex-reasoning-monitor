//go:build linux || darwin || windows

package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

//go:embed web/*
var dashboardFiles embed.FS

const dashboardAddress = "127.0.0.1:5927"

type radarSession struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Status        string    `json:"status"`
	Since         time.Time `json:"since"`
	Updated       time.Time `json:"updated"`
	Output        uint64    `json:"output"`
	Reasoning     uint64    `json:"reasoning"`
	Responses     uint64    `json:"responses"`
	Count516      uint64    `json:"count516"`
	LastOutput    uint64    `json:"lastOutput"`
	LastReasoning uint64    `json:"lastReasoning"`
	History       []uint64  `json:"history"`
}
type radarEvent struct {
	Sequence  uint64    `json:"sequence"`
	At        time.Time `json:"at"`
	Observed  time.Time `json:"observed"`
	Kind      string    `json:"kind"`
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Output    uint64    `json:"output"`
	Reasoning uint64    `json:"reasoning"`
}
type radarBucket struct {
	At        int64   `json:"at"`
	Output    uint64  `json:"output"`
	Reasoning uint64  `json:"reasoning"`
	Responses uint64  `json:"responses"`
	Count516  uint64  `json:"count516"`
	Covered   float64 `json:"covered"`
}
type radarSnapshot struct {
	Started   time.Time      `json:"started"`
	Now       time.Time      `json:"now"`
	Range     int            `json:"range"`
	Step      int64          `json:"step"`
	Output    uint64         `json:"output"`
	Reasoning uint64         `json:"reasoning"`
	Responses uint64         `json:"responses"`
	Count516  uint64         `json:"count516"`
	Median    float64        `json:"median"`
	P95       uint64         `json:"p95"`
	Covered   float64        `json:"covered"`
	Storage   string         `json:"storage"`
	Sessions  []radarSession `json:"sessions"`
	Events    []radarEvent   `json:"events"`
	Series    []radarBucket  `json:"series"`
	Peak30    *radarEvent    `json:"peak30"`
	PeakColor string         `json:"peakColor"`
}
type historyEntry struct {
	event radarEvent
	run   string
}
type dashboard struct {
	mu           sync.Mutex
	started      time.Time
	now          func() time.Time
	runID        string
	sequence     uint64
	sessions     map[string]*radarSession
	store        *eventStore
	persistent   bool
	errOut       io.Writer
	storageError error
	stop         chan struct{}
	done         chan struct{}
	closeOnce    sync.Once
	closeError   error
}

func dashboardWithStore(store *eventStore) *dashboard {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		panic(err)
	}
	return &dashboard{started: time.Now(), now: time.Now, runID: hex.EncodeToString(random), sessions: map[string]*radarSession{}, store: store, errOut: io.Discard}
}
func newDashboard() *dashboard {
	store, err := openEventStore(":memory:")
	if err != nil {
		panic(err)
	}
	return dashboardWithStore(store)
}
func newPersistentDashboard(path string, errOut io.Writer) (*dashboard, error) {
	store, err := openEventStore(path)
	if err != nil {
		return nil, err
	}
	d := dashboardWithStore(store)
	d.errOut = errOut
	d.persistent = true
	d.record("MONITOR_START", "", "", d.now(), tokenCount{})
	if d.storageError != nil {
		store.db.Close()
		return nil, d.storageError
	}
	d.stop = make(chan struct{})
	d.done = make(chan struct{})
	go func() {
		defer close(d.done)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-d.stop:
				return
			case <-ticker.C:
				d.record("HEARTBEAT", "", "", d.now(), tokenCount{})
			}
		}
	}()
	return d, nil
}
func (d *dashboard) Close() error {
	if d == nil {
		return nil
	}
	d.closeOnce.Do(func() {
		if d.stop != nil {
			close(d.stop)
			<-d.done
			d.record("MONITOR_STOP", "", "", d.now(), tokenCount{})
		}
		d.mu.Lock()
		defer d.mu.Unlock()
		d.closeError = d.store.db.Close()
		d.fail(d.closeError)
	})
	return d.closeError
}
func (d *dashboard) fail(err error) {
	if err != nil && d.storageError == nil {
		d.storageError = err
		fmt.Fprintf(d.errOut, "codex-reasoning-monitor: LOG DATABASE ERROR; check storage and restart: %v\n", err)
	}
}
func (d *dashboard) record(kind, id, title string, at time.Time, count tokenCount) {
	d.recordSource(kind, id, title, at, count, "", nil)
}
func (d *dashboard) recordSource(kind, id, title string, at time.Time, count tokenCount, eventID string, source *sourcePosition) bool {
	if d == nil {
		return true
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.storageError != nil {
		return false
	}
	now := d.now().Truncate(time.Millisecond)
	if at.IsZero() {
		at = now
	}
	d.sequence++
	if eventID == "" {
		eventID = fmt.Sprintf("%s:%d", d.runID, d.sequence)
	}
	e := radarEvent{At: at, Observed: now, Kind: kind, ID: id, Title: title, Output: count.output, Reasoning: count.reasoning}
	if title != "" {
		e.Title = sanitizeTitle(title)
	}
	inserted, err := d.store.append(&e, eventID, d.runID, source)
	if err != nil {
		d.fail(err)
		return false
	}
	if !inserted {
		return false
	}
	if id != "" {
		session := d.sessions[id]
		if session == nil {
			session = &radarSession{ID: id, Since: at, Status: "stopped", History: []uint64{}}
			d.sessions[id] = session
		}
		session.Updated = now
		if title != "" {
			session.Title = e.Title
		}
		switch kind {
		case "WAITING":
			session.Status = "waiting"
		case "START", "ATTACH":
			session.Status = "active"
		case "STOP":
			delete(d.sessions, id)
		}
	}
	return true
}
func (d *dashboard) resumePosition(id string) (sourcePosition, bool) {
	if d == nil {
		return sourcePosition{}, false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	p, ok, err := d.store.position(id)
	d.fail(err)
	return p, ok
}
func peakColor(e *radarEvent) string {
	if e == nil {
		return "none"
	}
	switch e.Reasoning {
	case 1000:
		return "green"
	case 516:
		return "red"
	default:
		return "yellow"
	}
}
func rangeStep(minutes int) int64 {
	switch minutes {
	case 15:
		return 10
	case 60:
		return 60
	case 1440:
		return 300
	case 10080:
		return 3600
	}
	return 0
}
func (d *dashboard) snapshot() radarSnapshot { return d.snapshotRange(60) }
func (d *dashboard) snapshotRange(minutes int) radarSnapshot {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now().Truncate(time.Millisecond)
	step := rangeStep(minutes)
	if step == 0 {
		minutes = 1440
		step = 300
	}
	cutoff := now.Add(-time.Duration(minutes) * time.Minute)
	history, err := d.store.historySince(cutoff.Add(-45*time.Second), now)
	d.fail(err)
	r := radarSnapshot{Started: d.started, Now: now, Range: minutes, Step: step, Storage: "memory", Sessions: []radarSession{}, Events: []radarEvent{}, Series: []radarBucket{}}
	if d.persistent {
		r.Storage = "ok"
	}
	if d.storageError != nil {
		r.Storage = "error"
	}
	// Exact rolling windows, with partially filled boundary buckets.
	first := cutoff.Unix() / step * step
	last := now.Unix() / step * step
	for stamp := first; stamp <= last; stamp += step {
		r.Series = append(r.Series, radarBucket{At: stamp})
	}
	sessions := map[string]*radarSession{}
	for id, s := range d.sessions {
		if s.Status != "stopped" {
			copy := *s
			copy.History = []uint64{}
			sessions[id] = &copy
		}
	}
	var values []uint64
	for _, h := range history {
		e := h.event
		if e.Observed.After(now) {
			continue
		}
		if e.Observed.Before(cutoff) || e.ID == "" {
			continue
		}
		r.Events = append(r.Events, e)
		s := sessions[e.ID]
		if s == nil {
			s = &radarSession{ID: e.ID, Since: e.At, Status: "stopped", History: []uint64{}}
			sessions[e.ID] = s
			if live := d.sessions[e.ID]; live != nil {
				s.Status = live.Status
			}
		}
		s.Updated = e.Observed
		if e.Title != "" {
			s.Title = e.Title
		}
		if e.Kind == "TOKENS" {
			s.Output += e.Output
			s.Reasoning += e.Reasoning
			s.Responses++
			s.LastOutput = e.Output
			s.LastReasoning = e.Reasoning
			s.History = append(s.History, e.Reasoning)
			if len(s.History) > 24 {
				s.History = s.History[len(s.History)-24:]
			}
			r.Output += e.Output
			r.Reasoning += e.Reasoning
			r.Responses++
			values = append(values, e.Reasoning)
			b := &r.Series[(e.Observed.Unix()/step*step-first)/step]
			b.Output += e.Output
			b.Reasoning += e.Reasoning
			b.Responses++
			if e.Reasoning == 516 {
				r.Count516++
				s.Count516++
				b.Count516++
			}
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	if len(values) > 0 {
		n := len(values)
		r.Median = float64(values[n/2])
		if n%2 == 0 {
			r.Median = float64(values[n/2-1])/2 + float64(values[n/2])/2
		}
		r.P95 = values[int(math.Ceil(float64(n)*.95))-1]
	}
	sort.SliceStable(r.Events, func(i, j int) bool { return r.Events[i].Sequence < r.Events[j].Sequence })
	if len(r.Events) > 200 {
		r.Events = r.Events[len(r.Events)-200:]
	}
	for _, s := range sessions {
		r.Sessions = append(r.Sessions, *s)
	}
	sort.Slice(r.Sessions, func(i, j int) bool {
		if r.Sessions[i].Updated.Equal(r.Sessions[j].Updated) {
			return r.Sessions[i].ID < r.Sessions[j].ID
		}
		return r.Sessions[i].Updated.After(r.Sessions[j].Updated)
	})
	stopped := 0
	kept := r.Sessions[:0]
	for _, s := range r.Sessions {
		if s.Status == "stopped" {
			stopped++
			if stopped > 100 {
				continue
			}
		}
		kept = append(kept, s)
	}
	r.Sessions = kept
	for _, interval := range d.coverage(history, now) {
		start, end := math.Max(interval[0], float64(cutoff.UnixMilli())/1000), math.Min(interval[1], float64(now.UnixMilli())/1000)
		if end <= start {
			continue
		}
		r.Covered += end - start
		for i := range r.Series {
			b := &r.Series[i]
			overlap := math.Min(end, float64(b.At+step)) - math.Max(start, float64(b.At))
			if overlap > 0 {
				b.Covered += overlap
			}
		}
	}
	r.Peak30, err = d.store.peak30(now)
	d.fail(err)
	if d.storageError != nil {
		r.Storage = "error"
	}
	r.PeakColor = peakColor(r.Peak30)
	return r
}
func (d *dashboard) coverage(history []historyEntry, now time.Time) [][2]float64 {
	type runPoint struct {
		last   float64
		index  int
		closed bool
	}
	runs := map[string]runPoint{}
	var intervals [][2]float64
	entries := append([]historyEntry{}, history...)
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].event.Observed.Before(entries[j].event.Observed) })
	for _, h := range entries {
		at := float64(h.event.Observed.UnixMilli()) / 1000
		if h.event.Observed.After(now) {
			continue
		}
		p, ok := runs[h.run]
		if !ok || p.closed || at-p.last > 45 {
			intervals = append(intervals, [2]float64{at, at})
			p.index = len(intervals) - 1
		} else {
			intervals[p.index][1] = at
		}
		p.last = at
		p.closed = h.event.Kind == "MONITOR_STOP"
		runs[h.run] = p
	}
	current := float64(now.UnixMilli()) / 1000
	if p, ok := runs[d.runID]; ok && !p.closed && d.storageError == nil && current-p.last <= 45 {
		intervals[p.index][1] = current
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i][0] < intervals[j][0] })
	var merged [][2]float64
	for _, v := range intervals {
		if len(merged) > 0 && v[0] <= merged[len(merged)-1][1] {
			merged[len(merged)-1][1] = math.Max(merged[len(merged)-1][1], v[1])
		} else {
			merged = append(merged, v)
		}
	}
	return merged
}

type eventPage struct {
	Events []radarEvent `json:"events"`
	Next   uint64       `json:"next"`
}

func (d *dashboard) eventPage(minutes int, before uint64, only516 bool, session string) (eventPage, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now().Truncate(time.Millisecond)
	page, err := d.store.eventPage(now.Add(-time.Duration(minutes)*time.Minute), now, before, only516, session)
	d.fail(err)
	return page, err
}

func (d *dashboard) handler() http.Handler {
	assets, _ := fs.Sub(dashboardFiles, "web")
	files := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reject DNS rebinding and cross-origin reads of local session metadata.
		if r.Host != dashboardAddress && r.Host != "localhost:5927" {
			http.Error(w, "invalid host", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host {
			http.Error(w, "invalid origin", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'")
		if r.URL.Path == "/api/snapshot" || r.URL.Path == "/api/events" {
			minutes := 1440
			if raw := r.URL.Query().Get("range"); raw != "" {
				value, err := strconv.Atoi(raw)
				if err != nil || rangeStep(value) == 0 {
					http.Error(w, "invalid range", http.StatusBadRequest)
					return
				}
				minutes = value
			}
			if r.URL.Path == "/api/events" {
				before, err := strconv.ParseUint(r.URL.Query().Get("before"), 10, 64)
				if r.URL.Query().Get("before") != "" && err != nil {
					http.Error(w, "invalid cursor", http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				if r.Method != http.MethodHead {
					page, err := d.eventPage(minutes, before, r.URL.Query().Get("reasoning") == "516", r.URL.Query().Get("session"))
					if err != nil {
						http.Error(w, "history unavailable", http.StatusServiceUnavailable)
						return
					}
					_ = json.NewEncoder(w).Encode(page)
				}
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if r.Method != http.MethodHead {
				_ = json.NewEncoder(w).Encode(d.snapshotRange(minutes))
			}
			return
		}
		files.ServeHTTP(w, r)
	})
}

func (d *dashboard) serve() (*http.Server, error) {
	listener, err := net.Listen("tcp4", dashboardAddress)
	if err != nil {
		return nil, err
	}
	server := &http.Server{Handler: d.handler(), ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	go server.Serve(listener)
	return server, nil
}
