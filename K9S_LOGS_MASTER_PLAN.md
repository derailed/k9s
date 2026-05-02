# K9s Logs Master Plan

Comprehensive fix plan for the longstanding log viewer issues in k9s, as catalogued
in [#3051 (comment)](https://github.com/derailed/k9s/issues/3051#issuecomment-3485172152).

**Verified against:** k9s HEAD at commit `0da2636c` (April 2026), source code read
of all key files, and web research of 20+ GitHub issues/PRs.

---

## Problem Summary

The k9s built-in log viewer has a cluster of interrelated bugs that have persisted
across versions since ~2020. They fall into **seven root-cause categories**:

| Category | Root Cause | Affected Issues |
|---|---|---|
| **A. Bracket escaping** | tview interprets `[text]` as color/style markup. Raw log content containing `[INFO]`, `[ERROR]`, timestamps, JSON, etc. is misinterpreted. `sanitizeEsc` regex is greedy and broken. Save paths don't sanitize at all. | #3051, #3099, #1398, #1164, #809, #137 |
| **B. Line dropping** | Channel buffer of only 50 + `default:` case in `readLogs()` silently drops lines when consumer is slow. | #1483, #1510 |
| **C. Tail/buffer limits** | Default `tail: 100` (hard-capped at `MaxLogThreshold = 5000`) hides older logs; filter operates only on the buffered subset. | #994, #993, #3364 |
| **D. Partial-line handling** | `ReadBytes('\n')` blocks until newline; lines without trailing `\n` are invisible during active tailing. | #1846 |
| **E. Panics / crashes** | `tview.TextView.Write()` can return `n > len(p)`, violating `io.Writer` contract. `bytes.Buffer.WriteTo` panics. Triggered by pressing `1` (head jump). | #3802 |
| **F. Filter / navigation UX** | `/` key doesn't always activate filter mode; time-range hotkeys (2-6) don't reliably override `sinceSeconds`; horizontal scroll moves 1 char at a time. | #3652, #3228, #3703, #1513 |
| **G. Structured/ANSI log rendering** | No built-in JSON pretty-printing; ANSI escape codes partially supported, causing garbage rendering with colored loggers. | #2998, #2148, #2212 |

These categories interact: bracket escaping bugs become worse with line wrapping;
filter-with-tail hides matches because the tail window is too small; save/copy
de-escaping is incomplete because `sanitizeEsc` regex is greedy.

### What has already been fixed (acknowledged)

| PR | Description | Status |
|---|---|---|
| #2751 (`66f57beb`, Aug 2024) | Added `tview.EscapeBytes()` call at log ingestion in `readLogs()` | Merged |
| #2830 (`64e46b61`, Aug 2024) | Updated tview dep to improve log escaping | Merged |
| #3503 (`c1d07ea6`, Sep 2025) | Added retry mechanism with exponential backoff for log streaming | Merged |
| #3580 (`2ed72068`, Sep 2025) | Improved log retry logic and error handling | Merged |
| #3557 (`e762cc6d`, Oct 2025) | Fixed UTF-8 character encoding issues in log search/highlighting | Merged |
| #3669 (`6f0ecdec`) | Added logs column lock feature | Merged |

| PR | Description | Status |
|---|---|---|
| #3945 | Fix `sanitizeEsc` regex + apply to save/copy paths | **NOT merged** (exists as PR, fixes #3051/#3099) |
| #3816 | Defensive `Write()` wrapper to prevent io.Writer panic | **NOT merged** (fixes #3802) |

---

## Current Architecture (Rendering Pipeline)

```
Kubernetes API (io.ReadCloser)
        │
        ▼
readLogs() ─── bufio.Reader.ReadBytes('\n')
        │      tview.EscapeBytes(bytes)  ◄── escapes [ ] to [[ ]] (added PR #2751)
        │
        ▼
dao.LogItem{Bytes, Pod, Container}
        │
        ▼
channel (dao.LogChan, buffer=50) ─── default: drop + warn  ◄── CATEGORY B
        │                              (logChannelBuffer=50, pod.go:49)
        ▼
model.Log.Append() ─── bounded buffer (Lines cap, default 100, max 5000)
        │                Shift() evicts oldest    ◄── CATEGORY C
        │
        ▼
model.Log.Notify() / fireLogBuffChanged()
        │  LogItem.Render(paint, showTime, buf)
        │     injects [gray::b], [color::] tags around already-escaped bytes
        │     does NOT re-escape l.Bytes (correct — they're pre-escaped)
        │  applyFilter() ─── regex on rendered bytes (includes tview tags!)
        │
        ▼
view.Log.Flush(lines)
        │  ansiWriter.Write(line) per line  ◄── tview.ANSIWriter
        │  (no batching — one Write per line)
        │
        ▼
Logger (embeds *tview.TextView)  ◄── NO Write() wrapper — CATEGORY E
        │  DynamicColors=true, Wrap=configurable
        │  internal wrap/paint engine parses [tag] sequences
        │  GetText(true) strips tags but leaves bracket artifacts  ◄── CATEGORY A
        │
        ▼
Terminal (tcell)
```

### Key source files (verified against HEAD `0da2636c`)

| File | Role |
|---|---|
| `internal/dao/pod.go:357-520` | `tailLogs()`, `readLogs()` — stream reader with retry, `tview.EscapeBytes`, channel buffer=50 |
| `internal/dao/log_item.go` | `LogItem` struct, `Render()` — injects tview color tags around pre-escaped bytes |
| `internal/dao/log_items.go` | `LogItems` collection, `Filter()`, `Render()`, `Lines()` |
| `internal/dao/log_options.go` | `LogOptions` — tail lines, since, container config, `ToPodLogOptions()` |
| `internal/model/log.go` | `Log` model — buffer management, `Append()`, `Notify()`, `applyFilter()`, listeners |
| `internal/config/logger.go` | `Logger` config — `DefaultLoggerTailCount=100`, `MaxLogThreshold=5000`, `DefaultSinceSeconds=-1` |
| `internal/view/log.go` | `Log` view — `Flush()`, `SaveCmd()` (**no sanitizeEsc!**), key bindings, ANSIWriter setup |
| `internal/view/logger.go` | `Logger` — tview.TextView wrapper, `DynamicColors(true)`, `saveCmd()` (**no sanitizeEsc!**), **no Write() wrapper** |
| `internal/view/log_indicator.go` | `LogIndicator` — status bar showing autoscroll, wrap, timestamp toggles |
| `internal/view/helpers.go:48-64` | `sanitizeEsc()` (**greedy broken regex**), `cpCmd()` — only copy uses sanitize |
| `internal/view/clipboard.go` | `clipboardWrite()` — OSC52 + native clipboard |
| `github.com/derailed/tview` (v0.8.5) | Fork of rivo/tview with `EscapeBytes`, `ANSIWriter` |

---

## Fix Plan

### Fix 1: Correct bracket escaping end-to-end (Category A)

**Issues resolved:** #3051, #3099, #1398, #1164, #809, #137

**Existing PR:** #3945 (NOT merged as of HEAD `0da2636c`) — partially addresses this.

**Problem detail (verified in source):**

1. `readLogs()` in `dao/pod.go:487` calls `tview.EscapeBytes(bytes)` which doubles
   brackets (`[` → `[[`, `]` → `]]`). This was added in PR #2751 (Aug 2024) and is
   correct for display.

2. `LogItem.Render()` in `dao/log_item.go:70-100` writes the already-escaped `l.Bytes`
   into a buffer alongside tview color tags like `[gray::b]`. The escaped brackets
   render correctly in tview's display engine.

3. **Bug 1 — greedy `sanitizeEsc` regex:** `view/helpers.go:48` defines:
   ```go
   var bracketRX = regexp.MustCompile(`\[(.+)\[\]`)
   ```
   The `.+` is greedy, so on a line like `[INFO[] some text [ERROR[]` it matches
   from the first `[` to the last `[]`, consuming everything in between. Only `cpCmd`
   (copy to clipboard) calls `sanitizeEsc` — the existing tests in `helpers_test.go:331-352`
   only cover trivial single-bracket cases, missing multi-sequence and edge cases.

4. **Bug 2 — `SaveCmd` doesn't sanitize:** `log.go:423` passes `l.logs.GetText(true)`
   directly to `saveData()` without calling `sanitizeEsc`. Saved log files contain
   escaped bracket artifacts like `[INFO[]`. Similarly, `logger.go:158` (`saveCmd`)
   passes `l.GetText(true)` to `saveYAML()` without sanitization.

5. **Bug 3 — wrapping with escaped brackets:** tview's internal word-wrap engine
   miscalculates string widths when escaped brackets are present at wrap points,
   causing garbled rendering (#3099). This is in the `derailed/tview` fork (v0.8.5).

**Fix steps:**

1. **Replace `sanitizeEsc` regex** with a proper non-greedy, character-class version:
   ```go
   // Current (broken) — helpers.go:48:
   var bracketRX = regexp.MustCompile(`\[(.+)\[\]`)
   
   // Fixed — use [^\[] to stop at the next bracket:
   var bracketRX = regexp.MustCompile(`\[([^\[]*)\[\]`)
   ```
   Add comprehensive test cases to `helpers_test.go` covering:
   - Multiple `[tag[]` sequences on one line: `"[INFO[] some [ERROR[]"` → `"[INFO] some [ERROR]"`
   - Empty brackets `"[]"` → `"[]"` (no change)
   - Nested/adjacent bracket patterns: `"[a[b[]c[]"` 
   - JSON with arrays from escaped log: `'{"a":[[1,2,3[]]}'`
   - Bracket at start/end of line
   - Lines with NO bracket escapes (passthrough)

2. **Apply `sanitizeEsc` in BOTH save paths:**
   - `view/log.go:423` — `SaveCmd`:
     ```go
     path, err := saveData(..., sanitizeEsc(l.logs.GetText(true)))
     ```
   - `view/logger.go:158` — `saveCmd`:
     ```go
     saveYAML(..., l.title, sanitizeEsc(l.GetText(true)))
     ```

3. **Fix wrapping with escaped brackets** — This requires a fix in the `derailed/tview`
   fork (v0.8.5). The `TextView` wrap-width calculation must account for escaped
   bracket sequences (`[[` counts as 1 visible char, not 2). Options:
   - Patch `tview.TextView.reindexBuffer()` to skip escaped brackets in width calc
   - Alternative: strip tview tags before measuring wrap width, re-inject after
   - This is the hardest sub-fix and requires a PR to `github.com/derailed/tview`

4. **Add round-trip integration test** — In `dao/log_item_test.go` (which already
   has tests using `tview.Escape`), add a test that:
   - Creates `LogItem` with bracket-heavy content: `[INFO] {"data":[1,2]} [ERROR]`
   - Renders through `LogItem.Render()` → asserts tview tags are correct
   - Feeds through `sanitizeEsc()` → asserts output matches original log text
   - Verifies with wrap enabled vs. disabled

**Files to modify:**
- `internal/view/helpers.go:48-52` — fix `bracketRX` regex
- `internal/view/helpers_test.go:331-352` — add comprehensive test cases
- `internal/view/log.go:423` — apply `sanitizeEsc` in `SaveCmd`
- `internal/view/logger.go:158` — apply `sanitizeEsc` in `saveCmd`
- `derailed/tview` (upstream fork v0.8.5) — fix wrap-width calculation

---

### Fix 2: Eliminate log line dropping under back-pressure (Category B)

**Issues resolved:** #1483, #1510

**Problem detail (verified in source):**

In `dao/pod.go:488-497`, the channel send uses a `default` case that drops lines
when the consumer is slow:

```go
select {
case <-ctx.Done():
    return streamCanceled
case out <- item:
default:
    // Avoid deadlock if consumer is too slow
    slog.Warn("Dropping log line due to slow consumer",
        slogs.Container, opts.Info(),
    )
}
```

The channel is created in `tailLogs()` at `pod.go:358` with:
```go
out := make(LogChan, logChannelBuffer)  // logChannelBuffer = 50 (pod.go:49)
```

The buffer was recently changed (PR #3503, Sep 2025 — commit message says "reduce
log channel buffer size to prevent drops" which is confusingly worded). A buffer of
50 means only 50 lines can be queued before the `default:` case kicks in and silently
discards lines. For bursty loggers producing 100+ lines/second, this is easily exceeded.

The consumer is the `model.Log.updateLogs()` goroutine which calls `Append()` then
conditionally calls `Notify()` (which fires UI updates via `QueueUpdateDraw`). The UI
draw is the bottleneck — it runs on the main thread.

**Fix steps:**

1. **Increase channel buffer** — Change `logChannelBuffer` from 50 to 500 or make
   it configurable. The memory cost is negligible (~500 pointers):
   ```go
   // pod.go:49 — current:
   logChannelBuffer = 50
   // proposed:
   logChannelBuffer = 500
   ```
   Alternatively, read from `LogOptions` or config so users can tune it.

2. **Add a short timeout instead of immediate drop** — Replace `default:` with a
   brief timeout (50-100ms) to give the consumer time to catch up during UI draw:
   ```go
   select {
   case <-ctx.Done():
       return streamCanceled
   case out <- item:
   case <-time.After(50 * time.Millisecond):
       slog.Warn("Dropping log line due to slow consumer", ...)
   }
   ```
   Note: `time.After` allocates a timer per call. For the hot path, use a reusable
   `time.Timer` instead to avoid GC pressure.

3. **Surface drop count to the user** — Track dropped lines via an atomic counter
   on `LogOptions` or a new field in `model.Log`. Display in `LogIndicator`:
   ```
   Autoscroll:On  Wrap:Off  Timestamps:On  [yellow]Dropped: 42[-]
   ```
   Add a `LogDropped(count int)` method to `LogsListener` interface, or pass via
   existing `LogFailed` with a special sentinel.

4. **Batch writes in Flush** — Currently `view/log.go:362-367` writes line-by-line
   via `ansiWriter.Write()` inside `QueueUpdateDraw`. Batching reduces lock/draw
   contention. The flush already receives `[][]byte` — join them:
   ```go
   func (l *Log) Flush(lines [][]byte) {
       ...
       buf := bytes.Join(lines, nil)
       _, _ = l.ansiWriter.Write(buf)
       ...
   }
   ```

**Files to modify:**
- `internal/dao/pod.go:49` — increase `logChannelBuffer`
- `internal/dao/pod.go:488-497` — add timeout instead of `default:`
- `internal/model/log.go` — track drop count, new listener method
- `internal/view/log.go:362-367` — batch `Flush` writes
- `internal/view/log_indicator.go` — display drop warning

---

### Fix 3: Improve tail/buffer behavior and filter interaction (Category C)

**Issues resolved:** #994, #993, #3364

**Problem detail (verified in source):**

1. Default `tail: 100` (`config/logger.go:8: DefaultLoggerTailCount = 100`). Only
   the last 100 lines are fetched from the K8s API. The config `Validate()` method
   clamps `TailCount` to `[1, MaxLogThreshold(5000)]` — so `-1` (fetch all) is NOT
   allowed through the config path. Users expect to see all logs but only see a tiny
   window.

2. `DefaultSinceSeconds = -1` (tail all time, `logger.go:14`) is correct, but the
   in-app time-range hotkeys (2-6 in `view/log.go:252-256`) map to fixed `sinceCmd(n)`
   calls that restart the stream. Issue #3228 reports these don't reliably override
   the previous state, leaving the user stuck at ~3 minutes.

3. Filter operates client-side on the buffered subset via `model.Log.applyFilter()`
   at `model/log.go:335`. If a match exists outside the `tail: 100` window, it's
   invisible. This is confusing because `kubectl logs | grep` finds matches k9s can't.

4. `SaveCmd` (`log.go:423`) only saves the in-memory buffer — whatever `GetText(true)`
   returns, not the full K8s log history.

**Fix steps:**

1. **Increase default tail** — Change `DefaultLoggerTailCount` from 100 to 1000:
   ```go
   // internal/config/logger.go:8
   DefaultLoggerTailCount = 1000  // was 100
   ```
   This is a backward-compatible change. The `MaxLogThreshold = 5000` cap remains.
   Users can still set any value from 1-5000 in their config.

2. **Raise or remove `MaxLogThreshold`** — 5000 is low for modern pods. Consider
   raising to 50,000 or removing the hard cap. The real constraint is memory, and
   5000 lines is only ~500KB-2MB depending on line length:
   ```go
   MaxLogThreshold = 50_000  // was 5_000
   ```
   Guard with a warning in `Validate()` if very high values are set.

3. **Show tail limit in title bar** — The title already shows "tail"/"head"/time
   duration (`log.go:307-318`). Append the line count so users know their window:
   ```
   Logs (ns/pod)[tail:1000]  vs. current:  Logs (ns/pod)[tail]
   ```

4. **Filter-aware tail** — When a filter is first applied, automatically re-fetch
   with a larger window to avoid hiding matches. In `model/log.go:199-206`:
   ```go
   func (l *Log) Filter(q string) {
       l.mx.Lock()
       l.filter = q
       l.mx.Unlock()
       // If filter active and tail is limited, expand to max
       if q != "" && l.logOptions.Lines < MaxLogThreshold {
           l.logOptions.Lines = MaxLogThreshold
           l.Restart(ctx)
           return
       }
       l.fireLogCleared()
       l.fireLogBuffChanged(0)
   }
   ```
   Note: this requires passing `ctx` to `Filter()` or storing it.

5. **Save full logs option** — Add a `Ctrl+Shift+S` binding that uses the K8s API
   to stream all logs directly to a file, bypassing the in-memory buffer:
   ```go
   func (l *Log) SaveFullCmd(*tcell.EventKey) *tcell.EventKey {
       // Use dao accessor to call ToPodLogOptions with TailLines=nil (all)
       // Stream directly to file via io.Copy
       // Show progress in Flash
   }
   ```

**Files to modify:**
- `internal/config/logger.go:8,11` — increase `DefaultLoggerTailCount`, raise `MaxLogThreshold`
- `internal/config/logger_test.go` — update test expectations
- `internal/model/log.go:199-206` — filter-aware tail expansion
- `internal/view/log.go:307-338` — show line count in title, add SaveFullCmd
- `internal/view/log_indicator.go` — optionally show tail count

---

### Fix 4: Handle partial lines (no trailing newline) (Category D)

**Issues resolved:** #1846

**Problem detail (verified in source):**

In `dao/pod.go:484-485`, `r.ReadBytes('\n')` blocks until it sees a newline:
```go
r := bufio.NewReader(stream)
for {
    bytes, err := r.ReadBytes('\n')
```

If a container writes a partial line (e.g., a progress bar, a prompt, or a log
framework that buffers), the line is invisible until the next `\n` arrives.

The EOF handling at `pod.go:501-504` does emit partial lines on stream close:
```go
if errors.Is(err, io.EOF) {
    if len(bytes) > 0 {
        // Emit trailing partial line before EOF
        out <- opts.ToLogItem(tview.EscapeBytes(bytes))
    }
```
But during active tailing, partial lines remain stuck in the `bufio.Reader` buffer
indefinitely. This is particularly noticeable with:
- Python apps using `print(..., end='')` or `sys.stdout.write()` without flush
- Progress bars (tqdm, etc.)
- Interactive-style log output

**Fix steps:**

1. **Add a flush timeout for partial lines** — Replace the blocking `ReadBytes`
   with a goroutine-based timed read. When no `\n` is seen within 500ms, emit
   whatever has been accumulated:
   ```go
   func readLogs(ctx context.Context, stream io.ReadCloser, out chan<- *LogItem, opts *LogOptions) streamResult {
       r := bufio.NewReader(stream)
       for {
           // Use a goroutine + channel for non-blocking read
           ch := make(chan readResult, 1)
           go func() { bytes, err := r.ReadBytes('\n'); ch <- readResult{bytes, err} }()
           select {
           case <-ctx.Done():
               return streamCanceled
           case result := <-ch:
               // normal processing...
           case <-time.After(500 * time.Millisecond):
               // Check if bufio has anything buffered
               if r.Buffered() > 0 {
                   partial := make([]byte, r.Buffered())
                   r.Read(partial)
                   out <- opts.ToLogItem(tview.EscapeBytes(partial))
               }
           }
       }
   }
   ```
   Caution: this is tricky because `ReadBytes` is blocking. An alternative is to
   use `io.Reader` with `SetReadDeadline` if the underlying stream supports it,
   or use a custom `bufio.Scanner` with `SplitFunc`.

2. **Mark partial lines visually** — When emitting a partial line (no trailing `\n`),
   use a distinct rendering:
   ```go
   // In LogItem:
   IsPartial bool  // new field
   
   // In Render():
   if l.IsPartial {
       bb.WriteString("[gray::d]⏎[-::-]")
   }
   ```

**Files to modify:**
- `internal/dao/pod.go:472-520` — `readLogs()` with timed partial-line emission
- `internal/dao/log_item.go:16` — add `IsPartial` field
- `internal/dao/log_item.go:70` — handle `IsPartial` in `Render()`

---

### Fix 5: Prevent io.Writer panic on head jump (Category E)

**Issues resolved:** #3802

**Existing PR:** #3816 (NOT merged as of HEAD `0da2636c`)

**Problem detail (verified in source):**

`Logger` in `view/logger.go:17-24` embeds `*tview.TextView`:
```go
type Logger struct {
    *tview.TextView
    ...
}
```

`tview.TextView.Write()` can return `n > len(p)`, violating the `io.Writer` contract.
When `bytes.Buffer.WriteTo` (used internally by the log viewer during head-jump
operations) encounters this, it panics with "invalid Write count".

**Reproduction:** Open logs → press `1` (head) → panic.

There is NO `Write()` method on `Logger` currently — the embedded `TextView.Write()`
is called directly.

**Fix (single method, surgical):**

Add a defensive `Write` wrapper to `Logger` that clamps the return value:
```go
// internal/view/logger.go — add after the Logger struct definition:

// Write implements io.Writer while protecting against tview returning
// n > len(p), which violates the io.Writer contract.
func (l *Logger) Write(p []byte) (int, error) {
    n, err := l.TextView.Write(p)
    if n > len(p) {
        n = len(p)
    }
    return n, err
}
```

This is exactly what PR #3816 proposes. It shadows the embedded method and clamps
the count. No other changes needed.

**Files to modify:**
- `internal/view/logger.go` — add `Write()` method

---

### Fix 6: Fix filter key, time-range hotkeys, and log view UX (Category F)

**Issues resolved:** #3652, #3228, #3703, #1513

**Problem detail:**

These are four related UX issues in the log viewer:

1. **#3652 — `/` key doesn't activate filter consistently** — On some Linux terminal
   emulators, pressing `/` to enter filter mode doesn't register on the first press.
   Users report needing to press `Shift-/` twice. This is likely a `tcell` key event
   issue where `/` generates a different key code than expected in some terminals.

2. **#3228 — Time-range hotkeys don't reliably override** — The hotkeys 2-6
   (`view/log.go:252-256`) call `sinceCmd(n)` which sets `SinceSeconds` and calls
   `Restart()`. But the restart reuses the same `LogOptions` object, and under
   certain conditions the previous `SinceTime` field (a timestamp string) takes
   precedence over `SinceSeconds` in `ToPodLogOptions()` (`log_options.go:96-113`).

3. **#3703 — Horizontal scroll moves 1 char at a time** — When wrap is off, left/right
   arrows move the viewport by 1 character. Long JSON or stack trace lines require
   dozens of keypresses to navigate. Should jump ~8 characters per press.

4. **#1513 — "Waiting for logs" hang** — Long-running pods with large log histories
   cause indefinite hang on "Waiting for logs...". The default `sinceSeconds: -1`
   fetches ALL history, which can be GB of data for pods running for weeks/months.

**Fix steps:**

1. **Filter key fix** — In `Logger.bindKeys()` at `logger.go:72-79`, ensure the
   `/` key binding handles both `tcell.KeyRune` with rune `/` AND the dedicated
   slash key. Add a fallback in `keyboard()`:
   ```go
   func (l *Logger) keyboard(evt *tcell.EventKey) *tcell.EventKey {
       if a, ok := l.actions.Get(ui.AsKey(evt)); ok {
           return a.Action(evt)
       }
       // Fallback: if the rune is '/' and we're not already in cmd mode
       if evt.Rune() == '/' && !l.cmdBuff.InCmdMode() {
           return l.activateCmd(evt)
       }
       return evt
   }
   ```

2. **sinceSeconds hotkey fix** — In `sinceCmd()` at `log.go:381-395`, explicitly
   clear the `SinceTime` field when switching time ranges:
   ```go
   func (l *Log) sinceCmd(n int) func(...) *tcell.EventKey {
       return func(*tcell.EventKey) *tcell.EventKey {
           l.logs.Clear()
           l.model.logOptions.SinceTime = ""  // ← clear to prevent precedence conflict
           ctx := l.getContext()
           ...
       }
   }
   ```

3. **Faster horizontal scroll** — This likely requires changes in `derailed/tview`'s
   `TextView` or adding custom key handlers in `Logger` that call `ScrollTo` with
   larger offsets. Alternative: add `Home`/`End` key support for jumping to line
   start/end.

4. **"Waiting for logs" fix** — Change the default behavior to use `sinceSeconds: 300`
   (5 minutes) instead of `-1` for the initial fetch, then allow users to press `0`
   to load full history. This prevents the hang for long-running pods while keeping
   the existing "press 0 for all" escape hatch. Alternatively, add a timeout on the
   initial fetch with a progress indicator.

**Files to modify:**
- `internal/view/logger.go:82-88` — keyboard fallback for `/`
- `internal/view/log.go:381-395` — clear `SinceTime` in `sinceCmd`
- `internal/config/logger.go:14` — consider changing `DefaultSinceSeconds` from -1 to 300
- `derailed/tview` — horizontal scroll step size

---

### Fix 7: First-class external log viewer integration (Meta-fix)

**Issues resolved:** #1187 (and effectively works around ALL of the above)

**Problem detail:**

The comment in #3051 argues (correctly) that k9s should not try to be a full-featured
log viewer. Tools like `lnav`, `less`, `stern`, and `bat` handle rendering, filtering,
and searching far better. k9s should make it trivial to pipe logs to these tools.

**Current state (verified):** Plugins `log-stern.yaml` and `log-full.yaml` exist in
`plugins/` directory but:
- No `lnav` plugin ships by default
- `log-full.yaml` only provides `kubectl logs -f` and `kubectl logs | less` — no lnav
- `log-stern.yaml` uses `stern --tail 50` which has its own tail limit
- The plugin system requires manual setup (copy to `~/.config/k9s/plugins.yaml`)
- There's no built-in "open in external viewer" shortcut

**Fix steps:**

1. **Ship an `lnav` plugin** — Add `plugins/log-lnav.yaml`:
   ```yaml
   plugins:
     lnav-logs:
       shortCut: Ctrl-L
       confirm: false
       description: "Logs (lnav)"
       scopes:
         - pods
         - containers
       command: bash
       background: false
       args:
         - -c
         - |
           if [ -n "$CONTAINER" ]; then
             kubectl logs --follow --tail=5000 "$NAME" \
               -c "$CONTAINER" -n "$NAMESPACE" \
               --context "$CONTEXT" --kubeconfig "$KUBECONFIG" 2>&1
           else
             kubectl logs --follow --all-containers=true --tail=5000 "$NAME" \
               -n "$NAMESPACE" --context "$CONTEXT" \
               --kubeconfig "$KUBECONFIG" 2>&1
           fi | lnav
   ```

2. **Add a config option for default external log viewer** — Let users set a
   preferred viewer in their k9s config:
   ```yaml
   k9s:
     logger:
       externalViewer: lnav    # or "less", "bat", "stern"
       externalViewerKey: Ctrl-L
   ```
   When the key is pressed, k9s pipes raw `kubectl logs` output to the configured
   viewer, completely bypassing the tview rendering pipeline.

3. **Implement a built-in "raw pipe" command** — Add a key binding (e.g., `Ctrl-E`
   for "external") that:
   - Suspends the TUI (like vim does with `:!command`)
   - Runs `kubectl logs -f <pod> | $EXTERNAL_VIEWER`
   - Returns to k9s when the viewer exits
   This is more reliable than the plugin approach because it properly handles
   terminal state transitions.

**Files to modify:**
- `plugins/log-lnav.yaml` — new plugin file
- `internal/config/logger.go` — add external viewer config
- `internal/view/log.go` — add external viewer key binding
- `plugins/README.md` — document the new plugin

---

## Implementation Priority

| Priority | Fix | Effort | Impact | Issues Fixed |
|---|---|---|---|---|
| **P0** | Fix 5 (io.Writer panic) | Trivial (5 lines) | Critical — crashes k9s | #3802 |
| **P0** | Fix 7 (external viewer plugin) | Low (YAML file) | High — works around all issues | #1187 + all |
| **P0** | Fix 1 (bracket escaping) | Medium | High — affects every user with structured logs | #3051, #3099, #1398, #1164, #809 |
| **P1** | Fix 2 (line dropping) | Medium | High — silent data loss | #1483, #1510 |
| **P1** | Fix 3 (tail/buffer) | Medium | Medium — confusing UX, most common complaint | #994, #993, #3364 |
| **P1** | Fix 6 (filter/nav UX) | Medium | Medium — multiple paper-cuts | #3652, #3228, #3703, #1513 |
| **P2** | Fix 4 (partial lines) | Medium | Low — niche use case | #1846 |

Recommended execution order: **Fix 5 → Fix 7 → Fix 1 → Fix 2 → Fix 3 → Fix 6 → Fix 4**

Fix 5 (io.Writer panic) is a 5-line surgical fix that prevents a crash — ship first.
Fix 7 (external viewer plugin) is a single YAML file that immediately gives users a
reliable escape hatch. Fix 1 (bracket escaping) addresses the most-reported bug
family and has PR #3945 as a starting point. Fixes 2-4, 6 are incremental improvements.

### What can be submitted as PRs immediately

1. **Fix 5** — Cherry-pick or re-create PR #3816 (defensive `Logger.Write` wrapper)
2. **Fix 7** — Submit `plugins/log-lnav.yaml` as a new plugin PR
3. **Fix 1 (partial)** — Help land PR #3945 (bracket regex fix + save sanitization)
4. **Fix 3 (partial)** — Submit a PR to change `DefaultLoggerTailCount` from 100 to 1000

---

## Consolidated Issue Cross-Reference

### Category A — Bracket Escaping
| Issue | Title | Fix | Status |
|---|---|---|---|
| [#3051](https://github.com/derailed/k9s/issues/3051) | k9s interprets `[]` characters in a weird way | Fix 1 | Open/Stale, PR #3945 pending |
| [#3099](https://github.com/derailed/k9s/issues/3099) | Wrapped logs not rendered correctly with brackets | Fix 1 | Open/Stale, PR #3945 pending |
| [#1398](https://github.com/derailed/k9s/issues/1398) | Pod logs containing brackets not shown | Fix 1 | Open/Stale |
| [#1164](https://github.com/derailed/k9s/issues/1164) | Copy from container log alters copied string | Fix 1 | Open/Stale |
| [#809](https://github.com/derailed/k9s/issues/809) | Saving logs mangles closing square brackets | Fix 1 | Open/Stale |
| [#137](https://github.com/derailed/k9s/issues/137) | Log escaping (oldest report) | Fix 1 | Closed (partially) |

### Category B — Line Dropping
| Issue | Title | Fix | Status |
|---|---|---|---|
| [#1483](https://github.com/derailed/k9s/issues/1483) | Container log entries get dropped | Fix 2 | Open/Stale |
| [#1510](https://github.com/derailed/k9s/issues/1510) | Log entry is missing | Fix 2 | Open/Stale |

### Category C — Tail/Buffer Limits
| Issue | Title | Fix | Status |
|---|---|---|---|
| [#3364](https://github.com/derailed/k9s/issues/3364) | Logs missing with tail + filter | Fix 3 | Open/Stale |
| [#994](https://github.com/derailed/k9s/issues/994) | Oldest part of logs are not shown | Fix 3 | Open |
| [#993](https://github.com/derailed/k9s/issues/993) | Saves all log text (request for full save) | Fix 3 | Open |

### Category D — Partial Lines
| Issue | Title | Fix | Status |
|---|---|---|---|
| [#1846](https://github.com/derailed/k9s/issues/1846) | Pod logs not ending in line break not shown | Fix 4 | Open/Stale |

### Category E — Panics/Crashes
| Issue | Title | Fix | Status |
|---|---|---|---|
| [#3802](https://github.com/derailed/k9s/issues/3802) | Panic: invalid write count on head for logs | Fix 5 | Open, PR #3816 pending |

### Category F — Filter/Navigation UX
| Issue | Title | Fix | Status |
|---|---|---|---|
| [#3652](https://github.com/derailed/k9s/issues/3652) | Log filtering (`/`) sometimes doesn't work | Fix 6 | Open/Stale |
| [#3228](https://github.com/derailed/k9s/issues/3228) | Unable to view logs beyond ~3 minutes | Fix 6 | Open (labeled question) |
| [#3703](https://github.com/derailed/k9s/issues/3703) | Horizontal move in log view too slow | Fix 6 | Open/In-progress |
| [#1513](https://github.com/derailed/k9s/issues/1513) | "Waiting for logs" hang on long-running pods | Fix 6 | Open/Stale |

### Category G — Structured/ANSI Rendering
| Issue | Title | Fix | Status |
|---|---|---|---|
| [#2998](https://github.com/derailed/k9s/issues/2998) | Colored JSON logs | Fix 7 (workaround) | Open |
| [#2148](https://github.com/derailed/k9s/issues/2148) | Container JSON logs are barely readable | Fix 7 (workaround) | Open |
| [#2212](https://github.com/derailed/k9s/issues/2212) | Console log colors not working properly | Fix 7 (workaround) | Open |

### Meta
| Issue | Title | Fix | Status |
|---|---|---|---|
| [#1187](https://github.com/derailed/k9s/issues/1187) | Filter logs through lnav | Fix 7 | Open/Stale |

---

## Testing Strategy

### Existing test infrastructure (verified)

| Test file | Coverage |
|---|---|
| `internal/dao/log_item_test.go` | `LogItem.Render()` with `tview.Escape` — has basic bracket tests |
| `internal/dao/log_items_test.go` | `LogItems.Filter()`, fuzzy/regex matching |
| `internal/view/helpers_test.go:331-352` | `sanitizeEsc()` — **only 3 trivial cases** (empty, empty-brackets, single tag) |
| `internal/config/logger_test.go` | `Logger` config defaults and validation |

### What each fix needs

| Fix | Unit Tests | Integration Tests |
|---|---|---|
| Fix 1 (brackets) | Expand `Test_sanitizeEsc` with multi-sequence, JSON, nested cases | Round-trip: `tview.Escape(raw)` → `LogItem.Render()` → `sanitizeEsc(GetText(true))` == `raw` |
| Fix 2 (dropping) | Benchmark: producer at 1000 lines/sec, verify no drops with buffer=500 | Stress test with `logChannelBuffer` variations |
| Fix 3 (tail/buffer) | Update `TestNewLogger` expectations for new defaults | Filter-with-tail: verify filter expands window automatically |
| Fix 4 (partial) | Test `readLogs` with mock stream that writes without `\n` | Verify partial line appears within 500ms |
| Fix 5 (panic) | Test `Logger.Write` returns `min(n, len(p))` | Press `1` (head) repeatedly without panic |
| Fix 6 (UX) | Test `sinceCmd` clears `SinceTime` | Manual: test `/` key in multiple terminals |
| Fix 7 (plugin) | N/A | Manual: install plugin, press shortcut, verify lnav opens |

### Manual verification test pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: k9s-log-test
spec:
  containers:
  - name: bracket-test
    image: busybox
    command: ["/bin/sh", "-c"]
    args:
    - |
      while true; do
        echo "[$(date -Iseconds)] [INFO] normal log line with [brackets]"
        echo '{"level":"info","msg":"json log","data":[1,2,3],"nested":{"key":"[value]"}}'
        echo "[ERROR] stack trace at [com.example.Class:42] caused by [NullPointerException]"
        printf "partial line without newline..."
        sleep 1
        echo " ...continued after 1s"
        # Burst of fast logs to trigger line dropping
        for i in $(seq 1 200); do echo "[DEBUG] burst line $i of 200"; done
        sleep 2
      done
  - name: ansi-test
    image: busybox
    command: ["/bin/sh", "-c"]
    args:
    - |
      while true; do
        printf '\033[32m[INFO]\033[0m green info message\n'
        printf '\033[31m[ERROR]\033[0m red error message\n'
        sleep 3
      done
```

### Verification checklist

For each fix, verify:
- [ ] Display renders correctly with wrap ON and OFF
- [ ] Copy (`c`) output matches original log text
- [ ] Save (`Ctrl+S`) output matches original log text  
- [ ] Filter (`/pattern`) finds all matching lines
- [ ] Head jump (`1`) doesn't panic
- [ ] Time range hotkeys (2-6) switch reliably
- [ ] High-volume burst doesn't silently drop lines (check for "Dropping log line" in k9s debug log via `k9s --logLevel debug`)
- [ ] External viewer plugin (lnav) opens with correct log stream
