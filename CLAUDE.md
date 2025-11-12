# K9s Codebase Architecture Guide

This document provides a comprehensive overview of the K9s codebase architecture to help future Claude instances understand the high-level design patterns and component interactions.

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Core Packages](#core-packages)
3. [Design Patterns](#design-patterns)
4. [Data Flow](#data-flow)
5. [Key Component Interactions](#key-component-interactions)
6. [Configuration System](#configuration-system)
7. [TUI Structure](#tui-structure)
8. [Kubernetes API Interactions](#kubernetes-api-interactions)

---

## High-Level Architecture

K9s is a terminal-based Kubernetes cluster management tool built with a multi-layered architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                        cmd/root.go                          │
│                    (Application Entry Point)                │
└────────────────────────┬────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
   ┌────▼────┐      ┌────▼────┐    ┌─────▼─────┐
   │ internal│      │ internal│    │ internal/ │
   │ config/ │      │ client/ │    │ watch/    │
   │         │      │         │    │           │
   └────┬────┘      └────┬────┘    └─────┬─────┘
        │                │               │
        │         ┌──────▼───────────────┘
        │         │
        │    ┌────▼──────────┐
        └───►│  internal/    │
             │  client::     │
             │  Connection   │
             └─────┬─────────┘
                   │
    ┌──────────────┼──────────────┐
    │              │              │
┌───▼──────┐  ┌───▼─────┐  ┌────▼─────┐
│ internal/│  │ internal│  │ internal/ │
│ view/    │  │ model/  │  │ dao/      │
│(Pages,  │  │(Table,  │  │(Accessors)│
│ App,    │  │ Stack)  │  │           │
│Browsers)│  └─────────┘  └─────┬─────┘
└──┬──────┘                      │
   │     ┌───────────────────────┘
   │     │
   │ ┌───▼──────────┐
   │ │ internal/   │
   │ │ render/     │
   │ │(Renderers)  │
   │ └──────────────┘
   │
┌──▼─────────────────┐
│  internal/ui/      │
│ (TUI Components:   │
│  Table, Pages,     │
│  Prompt, Menu)     │
└────────────────────┘
```

### Layer Breakdown

1. **Entry Point** (`cmd/root.go`)
   - Cobra CLI setup
   - Configuration loading (K8s flags, K9s config)
   - Application initialization

2. **Configuration Layer** (`internal/config/`)
   - K9s-specific configuration (skins, hotkeys, views)
   - Kubernetes configuration management
   - Settings persistence

3. **Kubernetes Connectivity** (`internal/client/`, `internal/watch/`)
   - Kubernetes API client wrapper
   - Watch factory for resource informers
   - Connection management

4. **Data Access Layer** (`internal/dao/`)
   - Resource accessors (pods, deployments, etc.)
   - Generic and specialized DAOs
   - Data retrieval from Kubernetes

5. **Model Layer** (`internal/model/`, `internal/model1/`)
   - Table models for rendering tabular data
   - Stack management for view navigation
   - Event listeners and notifications

6. **View Layer** (`internal/view/`)
   - High-level UI components (Browser, Pod, Deployment views)
   - Page stack management
   - Resource viewer implementations

7. **UI Layer** (`internal/ui/`)
   - Low-level TUI components (Table, Pages, Prompt)
   - Styling and rendering
   - Event handling

8. **Render Layer** (`internal/render/`)
   - Resource-specific renderers
   - Column customization
   - Data formatting

---

## Core Packages

### `cmd/`
**Entry point for the application**
- `root.go`: Main CLI command setup, configuration loading, app initialization
- `version.go`: Version command
- `info.go`: Info command

Key flow:
```
Execute() → run() → loadConfiguration() → NewApp() → Init() → Run()
```

### `internal/config/`
**Configuration management**
- `config.go`: Main Config struct holding K9s and Kubernetes settings
- `k9s.go`: K9s-specific configuration (UI, refresh rate, themes)
- `data/`: Kubernetes configuration data layer
- `json/`: JSON schema and validation

**Key interfaces:**
- `KubeSettings`: Abstraction for Kubernetes config access

### `internal/client/`
**Kubernetes API communication**
- `config.go`: K8s REST client configuration
- `connection.go`: Manages API server connection, informer factories
- `gvr.go`: GroupVersionResource definitions and helpers

**Key abstractions:**
- `Connection`: Interface for Kubernetes cluster communication
- `GVR`: Identifies resource types (e.g., v1/Pod)

### `internal/watch/`
**Dynamic resource informer factory**
- `factory.go`: Creates Kubernetes informers for resources
- `forwarders.go`: Port forwarding management

**Key patterns:**
- Uses Kubernetes DynamicSharedInformerFactory
- Lazy initialization of informers per namespace
- Caching and re-syncing mechanisms

### `internal/dao/`
**Data Access Objects - Resource-specific accessors**
- `accessor.go`: Registry pattern for DAO objects
- Specialized DAOs: `pod.go`, `deployment.go`, `node.go`, etc.
- `generic.go`: Generic accessor for custom resources
- `helpers.go`: Shared data retrieval utilities

**Key pattern:** Registry maps GVRs to specific DAO implementations

### `internal/model/`
**Data models and state management**
- `table.go`: Table model for tabular data display
- `stack.go`: Stack-based navigation for views
- `types.go`: Component and listener interfaces
- `flash.go`: Flash message model
- `history.go`: Command/filter history

**Key interfaces:**
- `Component`: Generic UI component interface
- `TableListener`: Observer for table model changes
- `StackListener`: Observer for stack navigation

### `internal/model1/`
**Table data structures**
- `table_data.go`: Core table data representation
- `row_event.go`: Row event definitions
- `header.go`: Column header management
- `color.go`: Row coloring logic

### `internal/view/`
**High-level UI views and controllers**
- `app.go`: Main application view (orchestrates all views)
- `browser.go`: Generic resource browser
- `pod.go`, `deployment.go`, etc.: Resource-specific views
- `table.go`: Table view wrapper
- `page_stack.go`: View stack management
- `registrar.go`: View registry (maps resources to viewers)
- `types.go`: View-related interfaces

**Key pattern:** Registry maps GVRs to view implementations

### `internal/ui/`
**Low-level TUI components (wrapping tview)**
- `app.go`: tview.Application wrapper
- `table.go`: tview.Table wrapper with model binding
- `pages.go`: tview.Pages stack management
- `prompt.go`: Command/filter input prompt
- `types.go`: UI-level interfaces
- `types.go`: Styling and configuration

**Responsibility:** Wrap tview primitives with business logic

### `internal/render/`
**Resource renderers for table display**
- `base.go`: Base renderer class
- Specialized renderers: `pod.go`, `deployment.go`, etc.
- `container.go`: Container display logic
- `cust_col.go`: Custom column support

**Pattern:** Each resource type has a renderer defining columns and formatting

---

## Design Patterns

### 1. Observer Pattern (Listeners)

Used extensively for event notification:

```
┌──────────────────┐
│   Model/Data     │
│   (e.g., Table)  │
└────────┬─────────┘
         │ notifies
         ▼
┌──────────────────────────┐
│   Listeners (Views)      │
│ - TableDataChanged()     │
│ - TableNoData()          │
│ - TableLoadFailed()      │
└──────────────────────────┘
```

**Key listener interfaces:**
- `TableListener`: For table model changes
- `StackListener`: For view stack operations
- `ResourceViewerListener`: For resource view changes
- `StylesChanged`: For skin/theme changes

**Implementation:**
```go
// In models
listeners []TableListener
func (t *Table) AddListener(l TableListener) { ... }
func (t *Table) notifyListeners(data *model1.TableData) { ... }
```

### 2. Registry Pattern

Maps abstract identifiers (GVRs) to concrete implementations:

**DAO Registry** (`internal/dao/accessor.go`):
```go
accessors = map[*client.GVR]Accessor{
    client.PodGVR:    &Pod{},
    client.DpGVR:     &Deployment{},
    // ... more mappings
}
```

**View Registry** (`internal/view/registrar.go`):
```go
MetaViewers = map[*client.GVR]MetaViewer{
    client.PodGVR:    MetaViewer{viewerFn: NewPod},
    client.DpGVR:     MetaViewer{viewerFn: NewDeployment},
    // ... more mappings
}
```

**Pattern benefits:**
- Extensibility: Add new resources without modifying core code
- Abstraction: Generic resource handling falls back to defaults
- Dynamic routing: URL-style navigation (e.g., `:pod` → Pod viewer)

### 3. MVC-like Pattern

Though not strict MVC, the architecture follows similar separation:

- **Model** (`internal/model/`, `internal/model1/`): Data structures and state
- **View** (`internal/view/`, `internal/ui/`): UI components and rendering
- **Controller**: Mixed into View layer (event handling, commands)

```
User Input → View (keyboard handler) → Model (update state) → Listeners (redraw View)
```

### 4. Factory Pattern

**Watch Factory** (`internal/watch/factory.go`):
- Creates Kubernetes informers on-demand
- Manages informer lifecycle (Start, Stop)
- Provides List/Get/Watch operations

```go
factory := watch.NewFactory(connection)
factory.Start(namespace)
defer factory.Terminate()
```

**Table Factory** (implicit):
```go
accessor := dao.AccessorFor(factory, gvr) // Creates or reuses accessor
```

### 5. Command/Interpreter Pattern

**Command parsing and execution** (`internal/view/cmd/`):
- User types commands prefixed with `:`
- Parser interprets command string
- Router dispatches to appropriate handler

### 6. Adapter Pattern

**DAO Accessor Interface**:
- Each DAO implements `Accessor` interface
- Generic accessor used for unmapped resource types
- Adapts specific resource types to common interface

```go
type Accessor interface {
    Init(Factory, *client.GVR)
    List(ctx, ns, filter) ([]runtime.Object, error)
    Get(ctx, fqn) (runtime.Object, error)
    // ... more methods
}
```

---

## Data Flow

### Startup Flow

```
1. cmd/root.go:Execute()
   └─> loadConfiguration()
       ├─> client.NewConfig() - K8s config
       ├─> config.NewConfig() - K9s config
       ├─> client.InitConnection() - Connect to K8s
       └─> k9sCfg.Load()/Save() - Persist config

2. view.NewApp(config)
   └─> ui.App (wraps tview.Application)
   └─> view.App (extends ui.App with content management)

3. app.Init()
   ├─> Content.Init() - Initialize page stack
   ├─> factory.Start() - Start K8s informers
   ├─> layout() - Create UI layout
   └─> ReloadStyles()

4. app.Run()
   └─> Navigate to default view
   └─> Start event loop (keyboard input, updates)
```

### View Navigation Flow

```
User types `:pod`
   ↓
Prompt.ParseCommand()
   ↓
Router.Route(":pod")
   ↓
Lookup GVR from command
   ↓
view.registrar.loadCustomViewers()
   ↓
Browser/SpecificViewer.Init()
   ├─> dao.AccessorFor() - Get data accessor
   ├─> model.Table.SetInstance() or .SetLabelSelector()
   └─> Start watching resources
   ↓
PageStack.Push(view) - Add to navigation stack
   ├─> view.Start()
   ├─> app.SetFocus(view)
   └─> Render table
```

### Data Update Flow

```
Kubernetes Event (via informer)
   ↓
watch.Factory detects change
   ↓
model.Table.refresh()
   ├─> Fetch data via dao.Accessor
   ├─> Format via render.Renderer
   └─> Create model1.TableData
   ↓
model.Table.notifyListeners(TableData)
   ↓
ui.Table.TableDataChanged(TableData)
   ├─> Update tview.Table
   ├─> Apply colors via ColorerFunc
   ├─> Update selection state
   └─> Queue redraw
   ↓
tview renders updated table on screen
```

---

## Key Component Interactions

### 1. App Orchestration (`internal/view/app.go`)

```
view.App
├─ Extends: ui.App (TUI application wrapper)
├─ Manages:
│  ├─ Content: PageStack (view stack)
│  ├─ factory: watch.Factory (K8s informers)
│  ├─ clusterModel: model.ClusterInfo (cluster status)
│  ├─ command: view.Command (command routing)
│  └─ cmdHistory, filterHistory
├─ Responsibilities:
│  ├─ Keyboard input routing
│  ├─ View navigation (push/pop pages)
│  ├─ Model refresh coordination
│  └─ Cluster status updates
└─ Lifecycle:
   ├─ Init() - Setup all components
   ├─ Run() - Event loop
   └─ Stop() - Cleanup
```

### 2. Table View Rendering (`internal/ui/table.go` → `internal/view/table.go`)

```
ui.Table (low-level TUI)
├─ Wraps: tview.Table
├─ Contains:
│  ├─ model.Table (data model)
│  ├─ SelectTable (selection tracking)
│  ├─ KeyActions (keyboard bindings)
│  └─ Styles (theming)
└─ Implements: TableListener interface

   │
   ↓

view.Table (high-level view)
├─ Wraps: ui.Table
├─ Adds:
│  ├─ App reference
│  ├─ EnterFunc (row double-click handler)
│  ├─ EnvFunc (environment for plugins)
│  └─ BindKeysFn (additional key bindings)
└─ Data flow:
   └─ Keyboard → App.keyboard() → Actions → Model update
```

### 3. Page Stack and Navigation

```
PageStack (view.page_stack.go)
├─ Wraps: ui.Pages + model.Stack
├─ Implements: StackListener
├─ Push(view) → Page added, view.Start() called
├─ Pop() → Page removed, view.Stop() called
└─ Stack operations notify listeners (ui.Pages)

   │
   ↓

ui.Pages (internal/ui/pages.go)
├─ Wraps: tview.Pages + model.Stack
├─ StackListener implementation
├─ Adds page to tview when pushed
├─ Removes page from tview when popped
└─ Switches displayed page on top change
```

### 4. Resource Data Pipeline

```
watch.Factory (K8s informers)
   │ notified of resource changes
   ↓
dao.Accessor.List()/Get()
   │ retrieves and processes resources
   ↓
render.Renderer
   │ formats for display (columns, data extraction)
   ↓
model1.TableData
   │ structured data for UI
   ↓
model.Table
   │ holds data, notifies listeners
   ↓
ui.Table
   │ renders to tview.Table
   ↓
Screen
```

### 5. Config System

```
config.Config
├─ K9s: K9s-specific settings
│  ├─ UI: UI preferences
│  ├─ RefreshRate: Poll interval
│  ├─ ReadOnly: R/O mode flag
│  ├─ activeContextName: Current cluster
│  └─ ... more settings
├─ conn: Kubernetes connection
├─ settings: KubeSettings (K8s config data)
└─ Lifecycle:
   ├─ Load(file) - From YAML/JSON
   ├─ Save(file) - To YAML/JSON
   └─ Refine() - Apply CLI flags

   File storage:
   ~/.config/k9s/config.yaml
   ~/.config/k9s/<context>/<cluster>/config.yaml  # Context-specific
```

---

## TUI Structure

### Component Hierarchy

```
┌──────────────────────────────────────┐
│      tview.Application               │
│       (ui.App wrapping)              │
└────────────────┬─────────────────────┘
                 │
                 ├─ Root: Pages (PageStack)
                 │  └─ Current page: Component
                 │     ├─ Browser/Table/Dialog
                 │     ├─ LiveView
                 │     └─ Custom views
                 │
                 └─ Global views (always visible):
                    ├─ Menu (top navigation)
                    ├─ Header/Logo
                    ├─ Crumbs (breadcrumb)
                    └─ Prompt (command input)
```

### Keyboard Event Flow

```
User presses key
   │
   ↓
tview.Application.Input
   │
   ↓
PageStack.SetInputCapture()
   │ or
   ↓
Current Component.SetInputCapture()
   │
   ├─ Check app-level actions (Ctrl+R, :, etc.)
   │  → app.keyboard()
   │
   ├─ Check component actions
   │  → component.keyboard()
   │
   └─ Check global key actions
      → KeyActions.Get(key)
      → execute KeyAction.Action(evt)
         (might modify model, trigger navigation, etc.)
```

### Rendering Pipeline

```
Component receives data change notification
   │
   ├─ model.TableDataChanged(data)
   │
   ├─ ui.Table.OnTableDataChanged(data)
   │  ├─ Iterate rows in data
   │  ├─ For each row:
   │  │  ├─ Apply ColorerFunc (determine row color)
   │  │  ├─ Format cell values
   │  │  └─ SetCell(row, col, cell)
   │  │
   │  └─ Redraw required
   │
   └─ tview redraws on next cycle
```

### View Types

1. **Browser** - Generic resource table view
   - Uses model.Table with watch updates
   - Displays paginated list of resources
   - Supports filtering, sorting, bulk operations

2. **Details/LiveView** - Resource detail view
   - Shows YAML, describe output, logs
   - Real-time updates
   - Search/filter within content

3. **Dialog** - Modal popup
   - Prompts for confirmation, input
   - Blocks underlying view
   - Modal forms for actions

4. **Table** - Specialized table view
   - Similar to Browser but for specific resources
   - May have custom columns, actions
   - Pod view, Deployment view, etc.

---

## Kubernetes API Interactions

### Connection Management (`internal/client/connection.go`)

```go
Connection interface {
    Config() *rest.Config
    ClientSet() kubernetes.Interface
    DynamicClient() dynamic.Interface
    RESTClient() rest.Interface
    ConnectionOK() bool
    CheckConnectivity() bool
    HasMetrics() bool
    ValidNamespaceNames() ([]string, error)
    // ... more methods
}
```

**Initialization:**
```
client.NewConfig(flags)
   │ wraps kubectl flags
   ↓
client.InitConnection()
   │ creates REST config
   │ creates k8s clientsets
   ↓
connection ready to use
```

### Watch Mechanism (`internal/watch/`)

```
Factory.Start(namespace)
   ├─ Creates DynamicSharedInformerFactory per namespace
   ├─ Starts factory (begins watching)
   └─ stopChan controls lifecycle

Factory.List(gvr, ns)
   ├─ Gets informer for resource type
   ├─ Returns cached objects from informer
   └─ Can wait for cache sync

Factory.Get(gvr, fqn)
   └─ Single resource retrieval

Factory.CanForResource(ns, gvr, verbs)
   └─ Checks RBAC access
```

### Accessor Data Retrieval

```
dao.AccessorFor(factory, gvr)
   ├─ Looks up GVR in registry
   ├─ Falls back to generic accessor
   └─ Returns initialized Accessor

Accessor.Init(factory, gvr)
   └─ Stores factory reference

Accessor.List(ctx, ns)
   ├─ Calls factory.List()
   ├─ Processes resources (filter, transform)
   └─ Returns []runtime.Object

Accessor.Get(ctx, fqn)
   ├─ Calls factory.Get()
   ├─ Processes single resource
   └─ Returns runtime.Object
```

### Metrics Integration

```
If cluster has metrics-server:
   └─ connection.HasMetrics() → true
      ├─ CanForResource includes metrics verbs
      ├─ Accessors can fetch metrics
      └─ Table displays CPU/Memory columns

Else:
   └─ connection.HasMetrics() → false
      ├─ Metrics columns hidden
      └─ Accessors skip metrics retrieval
```

---

## Configuration System

### Config File Structure

```
~/.config/k9s/
├─ config.yaml                  # Global K9s config
├─ clusters/
│  └─ <cluster>/
│     ├─ config.yaml           # Cluster-specific config
│     └─ contexts/
│        └─ <context>/
│           ├─ config.yaml     # Context-specific config
│           ├─ aliases.yaml
│           ├─ hotkeys.yaml
│           └─ plugins.yaml
├─ skins/                       # Themes
│  └─ *.yaml
├─ views/                       # Custom view definitions
│  └─ *.yaml
└─ plugins/                     # Plugin definitions
   └─ *.yaml
```

### Config Loading Hierarchy

```
1. Default values (hardcoded)
2. Global config (~/.config/k9s/config.yaml)
3. Cluster-specific config
4. Context-specific config
5. CLI flags (overrides all)

Settings checked in reverse order (CLI flag highest priority)
```

### Settings Types

**K9s Settings** (internal/config/k9s.go):
```go
- RefreshRate: How often to poll K8s
- ReadOnly: Prevent modifications
- UI: Font, icons, colors, etc.
- Logger: Logging configuration
- Thresholds: Alert thresholds
- ShellPod: Pod exec defaults
- ImageScans: Vulnerability scanning
- DefaultView: Starting view
```

**K8s Settings** (internal/client/):
```go
- KubeConfig path
- Current context
- Namespace
- Request timeout
- Impersonation settings
```

---

## Summary

K9s architecture emphasizes:

1. **Separation of Concerns**: Clear layering from K8s API to UI rendering
2. **Extensibility**: Registry patterns for adding resources and views
3. **Reactive Updates**: Observer pattern for real-time UI updates
4. **Abstraction**: Interfaces (Accessor, ResourceViewer, Component) hide implementation details
5. **Efficiency**: Kubernetes informer factories for efficient resource watching
6. **Configuration**: Hierarchical, context-aware configuration system
7. **Navigation**: Stack-based view management for intuitive UX

Key insight: **The flow is consistently: K8s events → Models → Listeners → UI updates**

