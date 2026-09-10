//go:build linux || darwin || windows

package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SQLite is the sole history store. Events and the matching rollout position
// commit together; event_id prevents replayed lines from being counted twice.
type eventStore struct{ db *sql.DB }

func openEventStore(path string) (*eventStore, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, err
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return nil, err
		}
		if err = file.Close(); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	fail := func(err error) (*eventStore, error) { db.Close(); return nil, err }
	if _, err = db.Exec(`PRAGMA busy_timeout=5000; PRAGMA journal_mode=WAL; PRAGMA synchronous=FULL;`); err != nil {
		return fail(err)
	}
	var version int
	if err = db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fail(err)
	}
	if version > 1 {
		return fail(fmt.Errorf("unsupported log database version %d", version))
	}
	if _, err = db.Exec(`
 CREATE TABLE IF NOT EXISTS events (
  seq INTEGER PRIMARY KEY AUTOINCREMENT,
  event_id TEXT NOT NULL UNIQUE,
  run_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  session_id TEXT NOT NULL,
  title TEXT NOT NULL,
  occurred_at INTEGER NOT NULL,
  observed_at INTEGER NOT NULL,
  output INTEGER NOT NULL CHECK(output>=0),
  reasoning INTEGER NOT NULL CHECK(reasoning>=0)
 );
 CREATE INDEX IF NOT EXISTS events_time ON events(observed_at);
 CREATE INDEX IF NOT EXISTS events_516 ON events(observed_at,seq) WHERE kind='TOKENS' AND reasoning=516;
 CREATE INDEX IF NOT EXISTS events_session ON events(session_id,observed_at);
 CREATE TABLE IF NOT EXISTS sources (
  session_id TEXT PRIMARY KEY,
  path TEXT NOT NULL,
  offset INTEGER NOT NULL,
  prefix_length INTEGER NOT NULL,
  prefix_hash TEXT NOT NULL
 );
 PRAGMA user_version=1;
 `); err != nil {
		return fail(err)
	}
	return &eventStore{db: db}, nil
}
func (s *eventStore) append(e *radarEvent, eventID, runID string, source *sourcePosition) (bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`INSERT INTO events(event_id,run_id,kind,session_id,title,occurred_at,observed_at,output,reasoning) VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(event_id) DO NOTHING`, eventID, runID, e.Kind, e.ID, e.Title, e.At.UnixMilli(), e.Observed.UnixMilli(), e.Output, e.Reasoning)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}
	seq, err := result.LastInsertId()
	if err != nil {
		return false, err
	}
	if source != nil {
		_, err = tx.Exec(`INSERT INTO sources(session_id,path,offset,prefix_length,prefix_hash) VALUES(?,?,?,?,?) ON CONFLICT(session_id) DO UPDATE SET path=excluded.path,offset=excluded.offset,prefix_length=excluded.prefix_length,prefix_hash=excluded.prefix_hash`, e.ID, source.Path, source.Offset, source.PrefixLength, source.PrefixHash)
		if err != nil {
			return false, err
		}
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	e.Sequence = uint64(seq)
	return true, nil
}
func (s *eventStore) position(id string) (sourcePosition, bool, error) {
	var p sourcePosition
	err := s.db.QueryRow(`SELECT path,offset,prefix_length,prefix_hash FROM sources WHERE session_id=?`, id).Scan(&p.Path, &p.Offset, &p.PrefixLength, &p.PrefixHash)
	if errors.Is(err, sql.ErrNoRows) {
		return p, false, nil
	}
	return p, err == nil, err
}

const eventColumns = `seq,occurred_at,observed_at,kind,session_id,title,output,reasoning`

func scanEvent(scanner interface{ Scan(...any) error }) (radarEvent, error) {
	var e radarEvent
	var at, observed int64
	err := scanner.Scan(&e.Sequence, &at, &observed, &e.Kind, &e.ID, &e.Title, &e.Output, &e.Reasoning)
	e.At = time.UnixMilli(at).UTC()
	e.Observed = time.UnixMilli(observed).UTC()
	return e, err
}
func (s *eventStore) peak30(now time.Time) (*radarEvent, error) {
	e, err := scanEvent(s.db.QueryRow(`SELECT `+eventColumns+` FROM events WHERE kind='TOKENS' AND observed_at>=? AND observed_at<=? ORDER BY reasoning DESC,observed_at DESC,seq DESC LIMIT 1`, now.Add(-30*time.Minute).UnixMilli(), now.UnixMilli()))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}
func (s *eventStore) historySince(cutoff, now time.Time) ([]historyEntry, error) {
	rows, err := s.db.Query(`SELECT `+eventColumns+`,run_id FROM events WHERE observed_at>=? AND observed_at<=? ORDER BY seq`, cutoff.UnixMilli(), now.UnixMilli())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	history := []historyEntry{}
	for rows.Next() {
		var h historyEntry
		var at, observed int64
		e := &h.event
		if err = rows.Scan(&e.Sequence, &at, &observed, &e.Kind, &e.ID, &e.Title, &e.Output, &e.Reasoning, &h.run); err != nil {
			return nil, err
		}
		e.At = time.UnixMilli(at).UTC()
		e.Observed = time.UnixMilli(observed).UTC()
		history = append(history, h)
	}
	return history, rows.Err()
}
func (s *eventStore) eventPage(cutoff, now time.Time, before uint64, only516 bool, session string) (eventPage, error) {
	query := `SELECT ` + eventColumns + ` FROM events WHERE observed_at>=? AND observed_at<=? AND session_id<>''`
	args := []any{cutoff.UnixMilli(), now.UnixMilli()}
	if before > 0 {
		query += ` AND seq<?`
		args = append(args, before)
	}
	if only516 {
		query += ` AND kind='TOKENS' AND reasoning=516`
	}
	if session != "" {
		query += ` AND session_id=?`
		args = append(args, session)
	}
	query += ` ORDER BY seq DESC LIMIT 101`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return eventPage{}, err
	}
	defer rows.Close()
	page := eventPage{Events: []radarEvent{}}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return page, err
		}
		if len(page.Events) == 100 {
			page.Next = page.Events[99].Sequence
			break
		}
		page.Events = append(page.Events, e)
	}
	return page, rows.Err()
}

type sourcePosition struct {
	Path         string `json:"path"`
	Offset       int64  `json:"offset"`
	PrefixLength int    `json:"prefix_length"`
	PrefixHash   string `json:"prefix_hash"`
}

func tokenEventID(id string, end int64, line []byte) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00%d\x00", id, end)
	h.Write(line)
	return hex.EncodeToString(h.Sum(nil))
}
func filePrefix(file *os.File, length int) (string, int, error) {
	data := make([]byte, length)
	n, err := file.ReadAt(data, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", 0, err
	}
	hash := sha256.Sum256(data[:n])
	return hex.EncodeToString(hash[:]), n, nil
}
func (p sourcePosition) matches(file *os.File) bool {
	info, err := file.Stat()
	if err != nil || p.Offset < 0 || info.Size() < p.Offset || p.PrefixLength < 1 || p.PrefixLength > 128 {
		return false
	}
	hash, n, err := filePrefix(file, p.PrefixLength)
	return err == nil && n == p.PrefixLength && strings.EqualFold(hash, p.PrefixHash)
}
