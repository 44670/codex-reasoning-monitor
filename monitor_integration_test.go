//go:build linux || darwin || windows

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

const testID = "11111111-1111-4111-8111-111111111111"
const secondID = "22222222-2222-4222-8222-222222222222"

type capturedOutput struct {
	sync.Mutex
	buffer bytes.Buffer
}

func (o *capturedOutput) Write(data []byte) (int, error) {
	o.Lock()
	defer o.Unlock()
	return o.buffer.Write(data)
}

func (o *capturedOutput) text() string {
	o.Lock()
	defer o.Unlock()
	return o.buffer.String()
}

type integrationFixture struct {
	t    *testing.T
	home string
	db   *sql.DB
	out  capturedOutput
	err  capturedOutput
}

func fixture(t *testing.T) *integrationFixture {
	t.Helper()
	f := &integrationFixture{t: t, home: filepath.Join(t.TempDir(), "Codex 会话 #+%")}
	f.must(os.MkdirAll(filepath.Join(f.home, lockDirectoryName), 0700))
	var err error
	f.db, err = sql.Open("sqlite", filepath.Join(f.home, stateDatabaseName))
	f.must(err)
	_, err = f.db.Exec(`CREATE TABLE threads (id TEXT PRIMARY KEY, rollout_path TEXT, title TEXT, name TEXT)`)
	f.must(err)
	t.Cleanup(func() { f.db.Close() })
	return f
}

func (f *integrationFixture) must(err error) {
	f.t.Helper()
	if err != nil {
		f.t.Fatal(err)
	}
}

func (f *integrationFixture) rollout(id string) string { return filepath.Join(f.home, id+".jsonl") }
func (f *integrationFixture) lockPath(id string) string {
	return filepath.Join(f.home, lockDirectoryName, id+".lock")
}

func (f *integrationFixture) metadata(id, title string) {
	_, err := f.db.Exec(`INSERT INTO threads VALUES (?, ?, ?, NULL)`, id, f.rollout(id), title)
	f.must(err)
}

func (f *integrationFixture) append(id, data string) {
	file, err := os.OpenFile(f.rollout(id), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	f.must(err)
	_, err = file.WriteString(data)
	f.must(err)
	f.must(file.Close())
}

func (f *integrationFixture) lock(id string) *os.File {
	file, err := os.OpenFile(f.lockPath(id), os.O_CREATE|os.O_RDWR, 0600)
	f.must(err)
	f.must(acquireTestLock(file))
	f.t.Cleanup(func() { file.Close() })
	return file
}

func (f *integrationFixture) start(color string) {
	f.t.Helper()
	name := "codex-reasoning-monitor"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary, err := filepath.Abs(name)
	f.must(err)
	cmd := exec.Command(binary, "--codex-home", f.home, "--color", color, "--web=false", "--db", filepath.Join(f.home, "data", "log.sqlite"))
	cmd.Stdout, cmd.Stderr = &f.out, &f.err
	f.must(cmd.Start())
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	f.t.Cleanup(func() {
		_ = stopTestProcess(cmd.Process)
		select {
		case err := <-done:
			if err != nil && runtime.GOOS != "windows" {
				f.t.Errorf("monitor exit: %v; stderr: %s", err, f.err.text())
			}
		case <-time.After(3 * time.Second):
			_ = cmd.Process.Kill()
			<-done
			f.t.Error("monitor did not exit after stop request")
		}
	})
}

func (f *integrationFixture) wait(want string, timeout time.Duration) {
	f.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(f.out.text(), want) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	f.t.Fatalf("missing %q\nstdout:\n%sstderr:\n%s", want, f.out.text(), f.err.text())
}

func (f *integrationFixture) attached(color string) *os.File {
	f.metadata(testID, "测试会话")
	f.append(testID, tokenLine(999, 888))
	lock := f.lock(testID)
	f.start(color)
	f.wait("ATTACH", 3*time.Second)
	return lock
}

func tokenLine(output, reasoning int) string {
	return fmt.Sprintf(`{"timestamp":"2026-09-09T12:00:00.123Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"output_tokens":9999,"reasoning_output_tokens":9998},"last_token_usage":{"output_tokens":%d,"reasoning_output_tokens":%d}}}}`+"\n", output, reasoning)
}

func TestIntegrationLifecycleAndMultipleSessions(t *testing.T) {
	f := fixture(t)
	f.attached("never")
	f.metadata(secondID, "第二个会话")
	f.append(secondID, "")
	lock := f.lock(secondID)
	f.wait("START   "+secondID, 3*time.Second)
	f.append(testID, tokenLine(123, 45))
	f.append(secondID, tokenLine(456, 78))
	f.wait(`"测试会话"  output=123  reasoning=45`, 3*time.Second)
	f.wait(`"第二个会话"  output=456  reasoning=78`, 3*time.Second)
	f.must(lock.Close())
	f.must(os.Remove(f.lockPath(secondID)))
	f.wait("STOP    "+secondID, 3*time.Second)
	text := f.out.text()
	if strings.Contains(text, "output=999") || strings.Contains(text, "\x1b[") {
		t.Fatalf("history replay or unexpected ANSI: %q", text)
	}
	if !strings.Contains(text, displayTime(time.Date(2026, 9, 9, 12, 0, 0, 123000000, time.UTC))) {
		t.Fatal("token timestamp not preserved")
	}
}

func TestIntegrationPartialLineBurstAndOffset(t *testing.T) {
	f := fixture(t)
	f.attached("never")
	line := tokenLine(123, 45)
	f.append(testID, line[:len(line)-1])
	time.Sleep(100 * time.Millisecond)
	if strings.Contains(f.out.text(), "TOKENS") {
		t.Fatal("emitted unfinished line")
	}
	f.append(testID, "\n")
	f.wait("output=123", 3*time.Second)
	var burst strings.Builder
	for i := 1000; i < 1500; i++ {
		burst.WriteString(tokenLine(i, i-1000))
	}
	f.append(testID, burst.String())
	f.wait("output=1,499", 3*time.Second)
	time.Sleep(100 * time.Millisecond)
	if got := strings.Count(f.out.text(), "TOKENS"); got != 501 {
		t.Fatalf("expected exactly 501 events, got %d", got)
	}
}

func TestIntegrationLargeUnrelatedRecord(t *testing.T) {
	f := fixture(t)
	f.attached("never")
	quoted, err := json.Marshal(tokenLine(666, 666))
	f.must(err)
	f.append(testID, `{"type":"response_item","payload":{"text":`+string(quoted)+"}}\n")
	f.append(testID, `{"type":"response_item","text":"`+strings.Repeat("x", 2*maxBufferedLine)+"\"}\n"+tokenLine(321, 54))
	f.wait("output=321", 3*time.Second)
	if got := strings.Count(f.out.text(), "TOKENS"); got != 1 {
		t.Fatalf("expected one real token event, got %d", got)
	}
}

func TestIntegrationRenameTitleAndColor(t *testing.T) {
	f := fixture(t)
	f.attached("always")
	_, err := f.db.Exec(`UPDATE threads SET name = ? WHERE id = ?`, "PCB 新标题", testID)
	f.must(err)
	f.append(testID, tokenLine(777, 516))
	f.wait(ansiRed+"reasoning=516"+ansiReset, 3*time.Second)
	text := f.out.text()
	if !strings.Contains(text, `"PCB 新标题"`) || !strings.Contains(text, ansiCyan+"output=777"+ansiReset) {
		t.Fatalf("wrong title/color: %q", text)
	}
}

func TestIntegrationTruncateAndFinalDrain(t *testing.T) {
	f := fixture(t)
	lock := f.attached("never")
	f.must(os.Truncate(f.rollout(testID), 0))
	time.Sleep(100 * time.Millisecond)
	f.append(testID, tokenLine(17, 8))
	f.wait("output=17", 3*time.Second)
	f.append(testID, tokenLine(18, 9))
	f.must(lock.Close())
	f.must(os.Remove(f.lockPath(testID)))
	f.wait("STOP", 3*time.Second)
	text := f.out.text()
	if at := strings.Index(text, "output=18"); at < 0 || at > strings.Index(text, "STOP") {
		t.Fatalf("lost final event before STOP:\n%s", text)
	}
}

func TestIntegrationDelayedMetadataPreservesNewTokens(t *testing.T) {
	f := fixture(t)
	f.attached("never")
	f.append(secondID, "")
	f.lock(secondID)
	time.Sleep(100 * time.Millisecond)
	f.append(secondID, tokenLine(456, 78))
	f.metadata(secondID, "延迟落库")
	f.wait("START   "+secondID, 3*time.Second)
	f.wait("output=456", time.Second)
}

func TestIntegrationReplacementPreservesNewTokens(t *testing.T) {
	f := fixture(t)
	f.attached("never")
	f.must(os.Rename(f.rollout(testID), f.rollout(testID)+".old"))
	f.append(testID, tokenLine(654, 87))
	time.Sleep(500 * time.Millisecond)
	f.append(testID, tokenLine(655, 88))
	f.wait("output=655", 3*time.Second)
	f.wait("output=654", time.Second)
	if got := strings.Count(f.out.text(), "TOKENS"); got != 2 {
		t.Fatalf("expected exactly two events after replacement, got %d", got)
	}
}

func TestIntegrationRenameSameFilePreservesPartialLine(t *testing.T) {
	f := fixture(t)
	f.attached("never")
	line := tokenLine(456, 78)
	f.append(testID, line[:80])
	time.Sleep(100 * time.Millisecond)
	moved := f.rollout(testID) + ".moved"
	f.must(os.Rename(f.rollout(testID), moved))
	_, err := f.db.Exec(`UPDATE threads SET rollout_path = ? WHERE id = ?`, moved, testID)
	f.must(err)
	file, err := os.OpenFile(moved, os.O_APPEND|os.O_WRONLY, 0600)
	f.must(err)
	_, err = file.WriteString(line[80:])
	f.must(err)
	f.must(file.Close())
	f.wait("output=456", 3*time.Second)
	if got := strings.Count(f.out.text(), "TOKENS"); got != 1 {
		t.Fatalf("expected one event without replay after rename, got %d", got)
	}
}

func TestIntegrationResumeDoesNotReplayHistory(t *testing.T) {
	f := fixture(t)
	f.attached("never")
	f.metadata(secondID, "恢复会话")
	f.append(secondID, tokenLine(888, 777))
	f.lock(secondID)
	f.wait("START   "+secondID, 3*time.Second)
	f.append(secondID, tokenLine(456, 78))
	f.wait("output=456", 3*time.Second)
	if strings.Contains(f.out.text(), "output=888") {
		t.Fatalf("resumed history replayed:\n%s", f.out.text())
	}
}
