# codex-reasoning-monitor

Built for people who want to understand how deeply their own Codex is thinking.

The tool displays the reasoning and output token counts reported for each model response,
alongside the time and session title. Session IDs appear on START, ATTACH, and STOP lines.
This lets you compare how much reasoning Codex uses
across responses, including while it works autonomously through multiple tool calls.
Reasoning token counts indicate the amount of reported reasoning, not its quality or correctness.

Linux, Windows, and macOS stdout monitor for Codex sessions. It observes Codex writer-lock lifecycle,
resolves each session's rollout path and title from `state_5.sqlite`, then follows only the active
JSONL files from their current offsets.

Includes **Codex 本地智力雷达 / Codex Local Intelligence Radar**, a live browser dashboard at
**http://127.0.0.1:5927**. The dashboard starts with the monitor; console output continues as usual.
Startup prints a prominent bilingual dashboard banner with the URL (Ctrl/Cmd-click in terminals
that support links).
All HTML, CSS, and JavaScript are embedded in the executable. No frontend build, CDN, or additional
runtime dependencies are needed.

## Local dashboard / 本地看板

启动程序后，在浏览器打开 **http://127.0.0.1:5927**。界面根据浏览器首选语言自动选择中文或英文；
右上角的 **中 / EN** 可随时切换，选择会保存在当前浏览器中。

- 首屏仪表固定展示**最近 30 分钟内单次响应的最大推理 token 数**：恰好 **1000 为绿色**，
  恰好 **516 为红色**，其它数值为黄色；窗口内没有 token 事件时显示暂无数据。
  峰值随时间过期自动更新，不受历史范围、516 筛选或事件流暂停影响。
- 支持 **15m / 1h / 24h / 7d** 范围，展示范围内的推理量、输出量、响应次数、
  **516 事件次数与占比**、单次推理量中位数和 P95。点击 516 卡片可筛选对应事件与会话。
- 趋势图区分有监控记录与未覆盖的时段；会话支持搜索、状态筛选、详情与 ID 复制。
  事件列表支持暂停和加载更早记录。
- 历史自动存入**可执行文件所在目录下的 `data/log.sqlite`**，重启保留。
  已记录的源日志从持久化位置继续读取，事件与位置在同一事务提交，重读不会重复累计。
- 浏览器首选语言决定初始中英文，右上角可切换；演示数据有独立标记，不写入数据库。
- 推理 token 包含在报告的 output token 中，两者不相加；推理量反映报告的用量，
  不等于回答质量、智力评分或对内部思维的实时测量。

Open **http://127.0.0.1:5927** after starting the program. The UI defaults to Chinese when the
browser's preferred language is Chinese, and English otherwise. The **中 / EN** switch remembers
your choice in that browser.

The opening instrument shows the **maximum reasoning tokens in one response observed within the
last 30 minutes**: exactly **1000 is green**, exactly **516 is red**, and all other values are yellow.
It shows no data when the window is empty and recalculates as events expire. Historical range,
516 filtering, and feed pause do not affect this instrument.

Choose **15m, 1h, 24h, or 7d** for totals, token charts, exact-516 counts and percentages, median,
and P95 reasoning per response. Click the 516 card to filter matching events and sessions. The
feed supports older pages and pause; session cards offer search, status filters, and details.
Charts mark periods without monitoring coverage. Both desktop and mobile layouts are supported.
**Explore demo** displays clearly labeled sample data without writing it to history.

History is stored in **`data/log.sqlite` beside the executable**, using the existing SQLite driver.
It survives restart. Source positions and events commit together; stable event IDs prevent replay
from inflating totals. Range statistics use the monitor's observation time, including catch-up
reads; event rows retain the original source time. Output includes reasoning, so the two counters
must not be added. Token volume does not measure reasoning quality or correctness.

SQLite is the sole history store: no daily JSONL archives, secondary index files, or custom file
locks. The database keeps observations until you remove it; selecting 7d does not delete older
records. SQLite may create `log.sqlite-wal` and `log.sqlite-shm` alongside it while running. Stop
the monitor before copying the database for a simple backup. Override its location when needed:

```bash
./codex-reasoning-monitor --db /path/to/data/log.sqlite
```

Only the IPv4 loopback interface is bound. Session titles, IDs, and token counts are exposed to the
local browser; prompts, rollout contents, and filesystem paths are not sent. An occupied port is
reported at startup. Console-only operation still records history; use:

```bash
./codex-reasoning-monitor --web=false
```

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
2026-09-09 18:20:17.511  TOKENS  "Sync and package Codex"  output=371  reasoning=238
2026-09-09 18:21:03.904  STOP    019f729b-9af4-7771-82fd-00f2f8d5020b  "Sync and package Codex"
```

- `START` means a writer lock was created after the monitor started.
- `ATTACH` means the writer lock already existed when the monitor started.
- `STOP` means the writer lock file disappeared.
- The displayed title prefers an explicitly assigned thread name, then falls back to the stored
  generated title.
- Previously observed rollout files resume from their persisted complete-token-line position.
  Files without a saved position start at their recorded EOF; retries retain that baseline. If metadata is not available yet, the eventual file is read from zero to avoid
  losing events written during the wait (an old file discovered this way can replay history).
- Replaced files are read from zero. A renamed file with the same file identity retains its
  offset and any incomplete line.
- In terminal output, reasoning values `516`, `1,034`, and `1,552` are red; values above `1,000` are green; other
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
title/color output, read-only WAL access, SQLite restart recovery and deduplication, atomic
event/checkpoint rollback, rolling-window boundaries, exact-516 statistics, and history pagination. A native SQLite interoperability subtest runs when
the `sqlite3` command is installed, otherwise that subtest is skipped.

On Windows, set `$env:CGO_ENABLED = '0'`, build the `.exe`, then run `go test -v -count=1 ./...`.
Test cleanup uses forced child termination on Windows; interactive Ctrl+C and automatic console
coloring still need manual verification there.

To run the supplied tests without installing Go, change into the appropriate `dist/<os>-<arch>`
directory and run `./monitor.test -test.v` (PowerShell: `.\monitor.test.exe -test.v`). Keep the
program beside the test executable and use that directory as the working directory.
