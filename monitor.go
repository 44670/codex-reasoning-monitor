//go:build linux || darwin || windows

package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	_ "modernc.org/sqlite"
)

const (
	lockDirectoryName = "thread-writer-locks"
	stateDatabaseName = "state_5.sqlite"
	readBufferSize    = 64 * 1024
	maxBufferedLine   = 1024 * 1024
	pendingRetryEvery = 200 * time.Millisecond
	fullResyncEvery   = 30 * time.Second
)

type activationKind string

const (
	activationStart  activationKind = "START"
	activationAttach activationKind = "ATTACH"
)

type threadMetadata struct {
	rolloutPath string
	title       string
}

type pendingSession struct {
	kind        activationKind
	detectedAt  time.Time
	nextTry     time.Time
	attempts    uint
	lastWarn    time.Time
	initialFile os.FileInfo
}

type monitoredSession struct {
	id          string
	title       string
	rolloutPath string
	file        *os.File
	offset      int64
	partial     []byte
	discarding  bool
	identity    os.FileInfo
}

type monitor struct {
	lockDirectory string
	watcher       *fsnotify.Watcher
	database      *sql.DB
	printer       *printer
	errOut        io.Writer
	lockWatched   bool
	sessions      map[string]*monitoredSession
	pending       map[string]*pendingSession
	pathToSession map[string]string
	readBuffer    []byte
}

func newMonitor(codexHome string, printer *printer, errOut io.Writer) (*monitor, error) {
	info, err := os.Stat(codexHome)
	if err != nil {
		return nil, fmt.Errorf("open Codex home %s: %w", codexHome, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("Codex home is not a directory: %s", codexHome)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create fsnotify watcher: %w", err)
	}
	if err := watcher.Add(codexHome); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("watch Codex home %s: %w", codexHome, err)
	}

	stateDatabase := filepath.Join(codexHome, stateDatabaseName)
	database, err := sql.Open("sqlite", readOnlySQLiteDSN(stateDatabase))
	if err != nil {
		watcher.Close()
		return nil, fmt.Errorf("open Codex state database: %w", err)
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	return &monitor{
		lockDirectory: filepath.Join(codexHome, lockDirectoryName),
		watcher:       watcher,
		database:      database,
		printer:       printer,
		errOut:        errOut,
		sessions:      make(map[string]*monitoredSession),
		pending:       make(map[string]*pendingSession),
		pathToSession: make(map[string]string),
		readBuffer:    make([]byte, readBufferSize),
	}, nil
}

func (m *monitor) Close() {
	for _, session := range m.sessions {
		if session.file != nil {
			session.file.Close()
		}
	}
	m.database.Close()
	m.watcher.Close()
}

func (m *monitor) Run(ctx context.Context) error {
	if err := m.installLockWatch(activationAttach); err != nil {
		return err
	}

	retryTicker := time.NewTicker(pendingRetryEvery)
	defer retryTicker.Stop()
	resyncTicker := time.NewTicker(fullResyncEvery)
	defer resyncTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-m.watcher.Events:
			if !ok {
				return errors.New("fsnotify event channel closed")
			}
			m.handleEvent(event)
		case watcherError, ok := <-m.watcher.Errors:
			if !ok {
				return errors.New("fsnotify error channel closed")
			}
			if errors.Is(watcherError, fsnotify.ErrEventOverflow) {
				fmt.Fprintln(m.errOut, "codex-reasoning-monitor: fsnotify queue overflow; resynchronizing")
				m.resynchronize()
				continue
			}
			fmt.Fprintf(m.errOut, "codex-reasoning-monitor: fsnotify: %v\n", watcherError)
		case now := <-retryTicker.C:
			m.retryPending(now)
			m.retryDetachedRollouts()
		case <-resyncTicker.C:
			m.resynchronize()
		}
	}
}

func (m *monitor) handleEvent(event fsnotify.Event) {
	path := filepath.Clean(event.Name)
	if path == m.lockDirectory {
		if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
			m.lockWatched = false
			m.stopAllSessions()
			return
		}
		if event.Op&fsnotify.Create != 0 {
			if err := m.installLockWatch(activationStart); err != nil {
				fmt.Fprintf(m.errOut, "codex-reasoning-monitor: %v\n", err)
			}
		}
		return
	}

	if filepath.Dir(path) == m.lockDirectory {
		id, ok := threadIDFromLockPath(path)
		if !ok {
			return
		}
		if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
			m.deactivate(id, time.Now())
			return
		}
		if event.Op&fsnotify.Create != 0 {
			m.queueActivation(id, activationStart, time.Now())
		}
		return
	}

	id, ok := m.pathToSession[path]
	if !ok {
		return
	}
	session := m.sessions[id]
	if session == nil {
		return
	}
	if event.Op&fsnotify.Write != 0 {
		m.readAvailable(session)
	}
	if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
		m.readAvailable(session)
		m.detachRollout(session)
	}
}

func (m *monitor) installLockWatch(existingKind activationKind) error {
	if m.lockWatched {
		return nil
	}
	info, err := os.Stat(m.lockDirectory)
	if errors.Is(err, os.ErrNotExist) {
		m.stopAllSessions()
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect lock directory %s: %w", m.lockDirectory, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("Codex writer lock path is not a directory: %s", m.lockDirectory)
	}
	if err := m.watcher.Add(m.lockDirectory); err != nil {
		return fmt.Errorf("watch lock directory %s: %w", m.lockDirectory, err)
	}
	m.lockWatched = true
	m.scanLocks(existingKind)
	return nil
}

func (m *monitor) scanLocks(missingKind activationKind) {
	entries, err := os.ReadDir(m.lockDirectory)
	if errors.Is(err, os.ErrNotExist) {
		m.lockWatched = false
		m.stopAllSessions()
		return
	}
	if err != nil {
		fmt.Fprintf(m.errOut, "codex-reasoning-monitor: scan writer locks: %v\n", err)
		return
	}

	seen := make(map[string]struct{}, len(entries))
	now := time.Now()
	for _, entry := range entries {
		id, ok := threadIDFromLockPath(entry.Name())
		if !ok {
			continue
		}
		seen[id] = struct{}{}
		if m.sessions[id] == nil && m.pending[id] == nil {
			m.queueActivation(id, missingKind, now)
		}
	}
	for id := range m.sessions {
		if _, ok := seen[id]; !ok {
			m.deactivate(id, now)
		}
	}
	for id := range m.pending {
		if _, ok := seen[id]; !ok {
			m.deactivate(id, now)
		}
	}
}

func (m *monitor) queueActivation(id string, kind activationKind, detectedAt time.Time) {
	if m.sessions[id] != nil || m.pending[id] != nil {
		return
	}
	pending := &pendingSession{
		kind:       kind,
		detectedAt: detectedAt,
		nextTry:    detectedAt,
	}
	// Capture the baseline once, not after metadata/watch retries. If no rollout can
	// be located yet, start at zero so the waiting period cannot silently lose tokens.
	if metadata, err := m.lookupThread(id); err == nil {
		pending.initialFile, _ = os.Stat(metadata.rolloutPath)
	}
	m.pending[id] = pending
	m.tryActivate(id, detectedAt)
}

func (m *monitor) tryActivate(id string, now time.Time) {
	pending := m.pending[id]
	if pending == nil || now.Before(pending.nextTry) {
		return
	}
	if _, err := os.Stat(filepath.Join(m.lockDirectory, id+".lock")); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			m.deactivate(id, now)
		}
		return
	}

	metadata, err := m.lookupThread(id)
	if err == nil {
		err = m.activate(id, metadata, pending)
	}
	if err == nil {
		delete(m.pending, id)
		return
	}

	pending.attempts++
	delay := pendingRetryEvery << min(pending.attempts, 3)
	if delay > 2*time.Second {
		delay = 2 * time.Second
	}
	pending.nextTry = now.Add(delay)
	if now.Sub(pending.detectedAt) >= 5*time.Second && now.Sub(pending.lastWarn) >= 30*time.Second {
		fmt.Fprintf(m.errOut, "codex-reasoning-monitor: waiting for session %s metadata: %v\n", id, err)
		pending.lastWarn = now
	}
}

func (m *monitor) activate(id string, metadata threadMetadata, pending *pendingSession) error {
	path := filepath.Clean(metadata.rolloutPath)
	file, err := openRollout(path)
	if err != nil {
		return fmt.Errorf("open rollout %s: %w", path, err)
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("stat rollout %s: %w", path, err)
	}
	if err := m.watcher.Add(path); err != nil {
		file.Close()
		return fmt.Errorf("watch rollout %s: %w", path, err)
	}

	session := &monitoredSession{
		id:          id,
		title:       metadata.title,
		rolloutPath: path,
		file:        file,
		identity:    info,
	}
	if pending.initialFile != nil && os.SameFile(pending.initialFile, info) {
		session.offset = pending.initialFile.Size()
	}
	m.sessions[id] = session
	m.pathToSession[path] = id
	m.printer.lifecycle(pending.detectedAt, string(pending.kind), id, session.title)
	// Catch appends that landed between the initial stat and installation of the watch.
	m.readAvailable(session)
	return nil
}

func (m *monitor) retryPending(now time.Time) {
	ids := make([]string, 0, len(m.pending))
	for id := range m.pending {
		ids = append(ids, id)
	}
	for _, id := range ids {
		m.tryActivate(id, now)
	}
}

func (m *monitor) retryDetachedRollouts() {
	for _, session := range m.sessions {
		if session.file != nil {
			continue
		}
		metadata, err := m.lookupThread(session.id)
		if err != nil {
			continue
		}
		session.title = metadata.title
		path := filepath.Clean(metadata.rolloutPath)
		file, err := openRollout(path)
		if err != nil {
			continue
		}
		info, err := file.Stat()
		if err != nil {
			file.Close()
			continue
		}
		if err := m.watcher.Add(path); err != nil {
			file.Close()
			continue
		}
		session.rolloutPath = path
		session.file = file
		if session.identity == nil || !os.SameFile(session.identity, info) {
			session.offset = 0
			session.partial = nil
			session.discarding = false
		}
		session.identity = info
		m.pathToSession[path] = session.id
		m.readAvailable(session)
	}
}

func (m *monitor) deactivate(id string, at time.Time) {
	if pending := m.pending[id]; pending != nil {
		metadata, err := m.lookupThread(id)
		title := ""
		if err == nil {
			title = metadata.title
		}
		m.printer.lifecycle(pending.detectedAt, string(pending.kind), id, title)
		m.printer.lifecycle(at, "STOP", id, title)
		delete(m.pending, id)
		return
	}

	session := m.sessions[id]
	if session == nil {
		return
	}
	m.readAvailable(session)
	if metadata, err := m.lookupThread(id); err == nil {
		session.title = metadata.title
	}
	m.printer.lifecycle(at, "STOP", id, session.title)
	m.detachRollout(session)
	delete(m.sessions, id)
}

func (m *monitor) detachRollout(session *monitoredSession) {
	if session.rolloutPath != "" {
		m.watcher.Remove(session.rolloutPath)
		delete(m.pathToSession, session.rolloutPath)
	}
	if session.file != nil {
		session.file.Close()
	}
	session.file = nil
	session.rolloutPath = ""
}

func (m *monitor) readAvailable(session *monitoredSession) {
	if session.file == nil {
		return
	}
	if info, err := session.file.Stat(); err == nil && info.Size() < session.offset {
		session.offset = 0
		session.partial = nil
		session.discarding = false
	}

	for {
		read, err := session.file.ReadAt(m.readBuffer, session.offset)
		if read > 0 {
			session.offset += int64(read)
			m.consumeChunk(session, m.readBuffer[:read])
		}
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			fmt.Fprintf(m.errOut, "codex-reasoning-monitor: read rollout for %s: %v\n", session.id, err)
			return
		}
		if read == 0 {
			return
		}
	}
}

func (m *monitor) consumeChunk(session *monitoredSession, chunk []byte) {
	if session.discarding {
		newline := bytes.IndexByte(chunk, '\n')
		if newline < 0 {
			return
		}
		session.discarding = false
		chunk = chunk[newline+1:]
	}

	if len(session.partial) > 0 {
		newline := bytes.IndexByte(chunk, '\n')
		if newline < 0 {
			if len(session.partial)+len(chunk) > maxBufferedLine {
				session.partial = nil
				session.discarding = true
				return
			}
			session.partial = append(session.partial, chunk...)
			return
		}
		if len(session.partial)+newline <= maxBufferedLine {
			session.partial = append(session.partial, chunk[:newline]...)
			m.processLine(session, session.partial)
		}
		session.partial = session.partial[:0]
		chunk = chunk[newline+1:]
	}

	for len(chunk) > 0 {
		newline := bytes.IndexByte(chunk, '\n')
		if newline < 0 {
			if len(chunk) > maxBufferedLine {
				session.discarding = true
				return
			}
			session.partial = append(session.partial, chunk...)
			return
		}
		if newline <= maxBufferedLine {
			m.processLine(session, chunk[:newline])
		}
		chunk = chunk[newline+1:]
	}
}

func (m *monitor) processLine(session *monitoredSession, line []byte) {
	line = bytesTrimTrailingCarriageReturn(line)
	count, ok := extractTokenCount(line)
	if !ok {
		return
	}
	if metadata, err := m.lookupThread(session.id); err == nil {
		session.title = metadata.title
	}
	m.printer.tokens(count.timestamp, session.id, session.title, count)
}

func (m *monitor) resynchronize() {
	if err := m.installLockWatch(activationStart); err != nil {
		fmt.Fprintf(m.errOut, "codex-reasoning-monitor: %v\n", err)
		return
	}
	if m.lockWatched {
		m.scanLocks(activationStart)
	}
	for _, session := range m.sessions {
		m.readAvailable(session)
	}
}

func (m *monitor) stopAllSessions() {
	now := time.Now()
	ids := make([]string, 0, len(m.sessions)+len(m.pending))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	for id := range m.pending {
		ids = append(ids, id)
	}
	for _, id := range ids {
		m.deactivate(id, now)
	}
}

func (m *monitor) lookupThread(id string) (threadMetadata, error) {
	var metadata threadMetadata
	err := m.database.QueryRow(
		`SELECT rollout_path, COALESCE(NULLIF(name, ''), NULLIF(title, ''), '') FROM threads WHERE id = ?`,
		id,
	).Scan(&metadata.rolloutPath, &metadata.title)
	if err != nil && strings.Contains(err.Error(), "no such column: name") {
		err = m.database.QueryRow(
			`SELECT rollout_path, COALESCE(NULLIF(title, ''), '') FROM threads WHERE id = ?`,
			id,
		).Scan(&metadata.rolloutPath, &metadata.title)
	}
	if err != nil {
		return threadMetadata{}, err
	}
	if metadata.rolloutPath == "" {
		return threadMetadata{}, errors.New("thread has no rollout path")
	}
	return metadata, nil
}

func readOnlySQLiteDSN(path string) string {
	uriPath := filepath.ToSlash(path)
	if filepath.VolumeName(path) != "" && !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath // Windows drive: file:///C:/... rather than file://C:/...
	}
	dsn := url.URL{Scheme: "file", Path: uriPath}
	query := dsn.Query()
	query.Set("mode", "ro")
	query.Add("_pragma", "busy_timeout(1000)")
	query.Add("_pragma", "query_only(1)")
	dsn.RawQuery = query.Encode()
	return dsn.String()
}

func threadIDFromLockPath(path string) (string, bool) {
	name := filepath.Base(path)
	id, ok := strings.CutSuffix(name, ".lock")
	if !ok || len(id) != 36 {
		return "", false
	}
	for index, character := range id {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			if character != '-' {
				return "", false
			}
			continue
		}
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return "", false
		}
	}
	return id, true
}

func bytesTrimTrailingCarriageReturn(line []byte) []byte {
	if len(line) > 0 && line[len(line)-1] == '\r' {
		return line[:len(line)-1]
	}
	return line
}
