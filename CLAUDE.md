# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## About the Project

K9s is a terminal-based UI (TUI) for managing Kubernetes clusters. It provides real-time monitoring and management of K8s resources via an interactive curses-style interface.

## Build & Development Commands

```bash
make build       # Build the binary to execs/k9s
make test        # Run all tests (clears cache first)
make cover       # Run tests with coverage (generates cov.out, opens HTML report)
```

Run tests directly with Go for more control:
```bash
go test ./...                          # All tests
go test ./internal/view/...            # Single package
go test -run TestFoo ./internal/dao/   # Single test
go test --count=1 ./...                # Skip cache
```

Lint configuration is in `.golangci.yml`. Key constraints: max line length 170, max cyclomatic complexity 35, max function statements 60.

## High-Level Architecture

The codebase has a clear layered architecture:

```
cmd/ → internal/view/ → internal/model/ → internal/dao/ → internal/client/
                ↓                ↓
          internal/ui/    internal/model1/
                ↓                ↓
          internal/render/ (transforms K8s objects → table rows)
```

### Key Packages

- **`cmd/`** — CLI entry point via Cobra. Parses flags, loads kubeconfig, calls `view/app.go`.
- **`internal/client/`** — Wraps `k8s.io/client-go`. Manages cluster connections, GVR (GroupVersionResource) handling, and metrics client.
- **`internal/dao/`** — Data Access Objects. Each K8s resource type has a specialized accessor (e.g., `pod.go`, `node.go`). All accessors register with a factory; unknown types fall back to a generic accessor. Handles Get/List/Watch, port forwarding, and log streaming.
- **`internal/render/`** — Transforms K8s API objects into `model1.Row` table rows. One renderer per resource type. Handles column specs, color-coding, and custom columns.
- **`internal/model/`** — UI state: table filtering/sorting, log buffering, navigation history, page stack. Bridges between dao and UI.
- **`internal/model1/`** — Pure data structures (rows, headers, table data, row events). Decoupled from UI — safe to use in tests without TUI setup.
- **`internal/ui/`** — TUI primitives using `derailed/tview` + `derailed/tcell`. Table widget, prompts, dialogs, menus.
- **`internal/view/`** — Composes ui + model into interactive resource browsers. `app.go` is the main controller; `browser.go` is the generic resource viewer; resource-specific views (e.g., `pod.go`, `log.go`, `exec.go`) extend it.
- **`internal/watch/`** — Kubernetes Informer integration for real-time resource change callbacks (OnAdd/OnUpdate/OnDelete).
- **`internal/config/`** — YAML config loading, themes, color definitions. Layered: defaults → `~/.config/k9s/config.yaml` → CLI flags.

### Adding a New Resource Type

Three files are typically needed:
1. **`internal/dao/<resource>.go`** — Accessor implementing resource-specific operations.
2. **`internal/render/<resource>.go`** — Renderer defining table columns and formatting.
3. **`internal/view/<resource>.go`** — View binding keyboard actions to resource operations.

Register the accessor in `internal/dao/accessor.go` and the renderer in `internal/model/registry.go`.

### Key Patterns

- **Informer-based real-time updates**: `watch.Factory` wraps K8s Informers; row events (add/update/delete) flow up through the model and trigger UI refresh.
- **GVR as resource identity**: Resources are identified by `GroupVersionResource` throughout — from the client layer to renderers.
- **PageStack navigation**: `model.Stack` manages the view history. Navigating into a resource pushes a new view; `<Escape>` pops back.
- **Action system**: User keypresses map to named actions registered per-view. Actions call dao methods, show confirmation dialogs, and refresh the table.
