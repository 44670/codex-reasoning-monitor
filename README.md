# codex-reasoning-monitor

Linux, Windows, and macOS stdout monitor for Codex sessions. It observes Codex writer-lock lifecycle,
resolves each session's rollout path and title from `state_5.sqlite`, then follows only the active
JSONL files from their current offsets.

It does not modify Codex, inject messages, or parse complete rollout records. Token events are
identified by their compact JSONL signature, and only the flat `last_token_usage` fields are
extracted.

## Build

Go 1.24 or newer is required. The SQLite driver is `modernc.org/sqlite` (pinned to v1.39.1
for Go 1.24 compatibility). No C compiler or system SQLite library is needed.

```bash
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -X main.version=0.1.0' -o codex-reasoning-monitor .
```

On Windows (PowerShell):

```powershell
$env:CGO_ENABLED = '0'
go build -trimpath -ldflags='-s -w -X main.version=0.1.0' -o codex-reasoning-monitor.exe .
```

Cross-compile from Linux/macOS by adding `GOOS=windows` or `GOOS=darwin`, and
`GOARCH=amd64` (Intel/AMD) or `GOARCH=arm64` (ARM/Apple Silicon). For example:

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags='-s -w -X main.version=0.1.0' -o codex-reasoning-monitor.exe .
```

Release artifacts are under `dist/<os>-<arch>/`, with `os` equal to `linux`, `windows`, or
`darwin`. Each directory contains `codex-reasoning-monitor` (with `.exe` on Windows) and a
compiled `monitor.test` test executable (also `.exe` on Windows). These generated files are ignored by Git.

Platform validation for this change:

| Platform | Architectures built with CGO disabled | Runtime validation |
| --- | --- | --- |
| Linux | amd64, arm64 | amd64: integration tests and `go vet` passed; arm64: compilation only |
| Windows | amd64, arm64 | Compilation only; user will verify on Windows |
| macOS | amd64, arm64 | Compilation only; user will verify on macOS |

Both the program and the test executable are cross-compiled for all six targets. Compilation
does not verify native filesystem notifications, Windows console behavior, or macOS permissions.

## Run

```bash
./codex-reasoning-monitor
```

On Windows use `.\codex-reasoning-monitor.exe`. Exit with Ctrl+C.

The default state directory is `$CODEX_HOME`, or the user's `.codex` directory when it is unset
(`$HOME/.codex` on Linux/macOS, `%USERPROFILE%\.codex` on Windows).

```bash
./codex-reasoning-monitor --codex-home /path/to/.codex
./codex-reasoning-monitor --color always
```

Color modes are `auto`, `always`, and `never`. `auto` enables ANSI colors when stdout is a terminal;
`NO_COLOR` disables them.
On Windows, automatic coloring enables the console's virtual-terminal mode when available,
restoring the previous mode on graceful exit. Use `--color always` to send ANSI through a pipe.
Rollout readers on Windows allow concurrent writing, renaming, and deletion via file-sharing flags.

Example output:

```text
2026-09-09 18:20:11.042  ATTACH  019f729b-9af4-7771-82fd-00f2f8d5020b  "Sync and package Codex"
2026-09-09 18:20:17.511  TOKENS  019f729b-9af4-7771-82fd-00f2f8d5020b  "Sync and package Codex"  output=371  reasoning=238
2026-09-09 18:21:03.904  STOP    019f729b-9af4-7771-82fd-00f2f8d5020b  "Sync and package Codex"
```

- `START` means a writer lock was created after the monitor started.
- `ATTACH` means the writer lock already existed when the monitor started.
- `STOP` means the writer lock file disappeared.
- The displayed title prefers an explicitly assigned thread name, then falls back to the stored
  generated title.
- Rollout files located on first detection start at their recorded EOF; retries retain that
  baseline. If metadata is not available yet, the eventual file is read from zero to avoid
  losing events written during the wait (an old file discovered this way can replay history).
- Replaced files are read from zero. A renamed file with the same file identity retains its
  offset and any incomplete line.
- Reasoning values `516`, `1,034`, and `1,552` are red; values above `1,000` are green; other
  reasoning values are yellow. Output values are cyan.

The monitor watches the Codex home directory, `thread-writer-locks`, and each active rollout.
Every 30 seconds it rescans the lock directory and catches up each rollout from its saved offset.
Session status follows lock file existence. If Codex crashes and leaves a lock file, the session
remains listed until Codex cleans up that file; `STOP` reflects cleanup time, not crash time.
An existing stale lock may also produce `ATTACH` when the monitor starts.
If the filesystem notification queue overflows, it performs the same reconciliation immediately.

## Integration tests

Build the executable using the command above, then run `CGO_ENABLED=0 go test -v -count=1 ./...`.
Tests start the actual executable against temporary SQLite databases, rollout files, and real
platform file locks. They cover lifecycle, offset/partial-line reading, rename/delete sharing,
title/color output, and read-only WAL access. A native SQLite interoperability subtest runs when
the `sqlite3` command is installed, otherwise that subtest is skipped.

On Windows, set `$env:CGO_ENABLED = '0'`, build the `.exe`, then run `go test -v -count=1 ./...`.
Test cleanup uses forced child termination on Windows; interactive Ctrl+C and automatic console
coloring still need manual verification there.

To run the supplied tests without installing Go, change into the appropriate `dist/<os>-<arch>`
directory and run `./monitor.test -test.v` (PowerShell: `.\monitor.test.exe -test.v`). Keep the
program beside the test executable and use that directory as the working directory.
