//go:build linux || darwin || windows

package main

import (
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSQLiteReadOnlyConnection(t *testing.T) {
	f := fixture(t)
	reader, err := sql.Open("sqlite", readOnlySQLiteDSN(filepath.Join(f.home, stateDatabaseName)))
	f.must(err)
	defer reader.Close()
	reader.SetMaxOpenConns(1)
	var queryOnly, busyTimeout int
	f.must(reader.QueryRow("PRAGMA query_only").Scan(&queryOnly))
	f.must(reader.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout))
	if queryOnly != 1 || busyTimeout != 1000 {
		t.Fatalf("connection settings: query_only=%d busy_timeout=%d", queryOnly, busyTimeout)
	}
	if _, err := reader.Exec("CREATE TABLE forbidden (id INTEGER)"); err == nil {
		t.Fatal("read-only connection accepted a write")
	}
	// mode=ro must still protect the file if query_only is disabled.
	_, err = reader.Exec("PRAGMA query_only=OFF")
	f.must(err)
	if _, err := reader.Exec("CREATE TABLE forbidden (id INTEGER)"); err == nil {
		t.Fatal("mode=ro accepted a write")
	}
}

func TestIntegrationSQLiteWALConcurrentWriter(t *testing.T) {
	f := fixture(t)
	f.db.SetMaxOpenConns(1)
	var journalMode string
	f.must(f.db.QueryRow("PRAGMA journal_mode=WAL").Scan(&journalMode))
	if journalMode != "wal" {
		t.Fatalf("journal_mode=%s", journalMode)
	}
	_, err := f.db.Exec("PRAGMA wal_autocheckpoint=0")
	f.must(err)
	f.attached("never")
	tx, err := f.db.Begin()
	f.must(err)
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE threads SET name='WAL committed' WHERE id=?", testID)
	f.must(err)
	f.append(testID, tokenLine(123, 45))
	f.wait(`"测试会话"  output=123  reasoning=45`, 3*time.Second)
	f.must(tx.Commit())
	f.append(testID, tokenLine(456, 78))
	f.wait(`"WAL committed"  output=456  reasoning=78`, 3*time.Second)
	info, err := os.Stat(filepath.Join(f.home, stateDatabaseName) + "-wal")
	f.must(err)
	if info.Size() <= 32 {
		t.Fatal("WAL has no frames")
	}
	if strings.Contains(f.err.text(), "locked") {
		t.Fatalf("WAL reader lock error: %s", f.err.text())
	}

	t.Run("NativeSQLiteWriter", func(t *testing.T) {
		cli, err := exec.LookPath("sqlite3")
		if err != nil {
			t.Skip("optional native SQLite interoperability check requires sqlite3 CLI")
		}
		cmd := exec.Command(cli, filepath.Join(f.home, stateDatabaseName), "UPDATE threads SET name='Native SQLite' WHERE id='"+testID+"';")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("native writer: %v: %s", err, output)
		}
		f.append(testID, tokenLine(789, 90))
		f.wait(`"Native SQLite"  output=789  reasoning=90`, 3*time.Second)
	})
}
