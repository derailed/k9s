# K9s Chat Feature

## Overview

This document outlines the comprehensive plan to integrate a chat feature into k9s using **k9s native UI components and patterns**. The chat will be accessible via the `:chat` command, displayed in a split panel on the right side, and will support LLM-powered assistance with shell command execution and context awareness.

## Objectives

1. **Enhanced User Experience**: Provide intelligent assistance for Kubernetes operations
2. **Context Awareness**: Chat understands current k9s state (namespace, resources, etc.)
3. **Safe Command Execution**: Shell commands with approval system
4. **Pure Markdown Rendering**: Using custom MarkdownTextView with Glamour for beautiful, consistent formatting
5. **Extensible Architecture**: Ready for future enhancements (history, multiple providers)
6. **Security**: Safe handling of API keys and command execution

## K9s Native Architecture Analysis

### Core UI Patterns in K9s

After analyzing k9s codebase, these are the key patterns we should follow:

#### 1. Component Structure (`internal/ui/`)
- **Base Pattern**: All UI components extend `tview` primitives
- **Styling**: Centralized through `config.Styles` with theme support
- **Actions**: Unified key binding system via `ui.KeyActions`
- **Examples**: `Table`, `Tree`, `Menu`, `Crumbs`, `Prompt`

#### 2. View Layer (`internal/view/`)
- **Component Interface**: Implement `model.Component` for integration
- **Initialization**: Standard `Init(context.Context) error` pattern
- **Key Handling**: Consistent keyboard event handling
- **Examples**: `Help`, `LiveView`, `Table` views

#### 3. Layout Management
- **Flex Containers**: `tview.Flex` for responsive layouts
- **Page Stack**: `ui.Pages` for view management
- **Border & Styling**: Consistent border and padding patterns

#### 4. State Management
- **App Context**: Central app state in `view.App`
- **Components**: State isolation in component structs
- **Events**: Listener pattern for style/state changes

## Technical Architecture (K9s Native)

### Core Components

#### 1. Chat Panel Integration
- **Split Layout Manager**: Reuse `tview.Flex` patterns from k9s
- **Toggle Command**: `:chat` command integration (already implemented)
- **Focus Management**: Tab-based focus switching between panels
- **Native Styling**: Full integration with k9s theme system

#### 2. Chat UI Components (Custom Markdown Stack)
```go
// Custom markdown-capable components with k9s integration
internal/chat/
├── component.go          # model.Component implementation
├── chat_view.go         # Main chat UI using tview.Flex
├── markdown_view.go     # Custom MarkdownTextView with Glamour rendering
├── types.go             # Message data structures and interfaces
├── mock_provider.go     # Mock LLM provider for testing
└── provider/            # Future LLM provider implementations
    ├── interface.go
    ├── openai.go
    └── mock.go
```

#### 3. Native UI Patterns to Follow

**Component Structure** (Following `internal/view/help.go` pattern):
```go
type Chat struct {
    *tview.Flex
    
    messageView *tview.TextView
    inputField  *tview.InputField
    actions     *ui.KeyActions
    app         *App
    styles      *config.Styles
    // ... other fields
}

func NewChat(app *App) *Chat {
    c := Chat{
        Flex:        tview.NewFlex(),
        messageView: tview.NewTextView(),
        inputField:  tview.NewInputField(),
        actions:     ui.NewKeyActions(),
        app:         app,
    }
    return &c
}

func (c *Chat) Init(ctx context.Context) error {
    c.SetBorder(true)
    c.SetBorderPadding(0, 0, 1, 1)
    c.bindKeys()
    c.setupLayout()
    c.app.Styles.AddListener(c)
    c.StylesChanged(c.app.Styles)
    return nil
}

// Implement model.Component interface
func (c *Chat) Name() string { return "chat" }
func (c *Chat) Hints() model.MenuHints { return c.actions.Hints() }
// ... etc
```

**Styling Pattern** (Following k9s native components):
```go
func (c *Chat) StylesChanged(s *config.Styles) {
    c.styles = s
    c.SetBackgroundColor(s.BgColor())
    c.SetBorderColor(s.Frame().Border.FgColor.Color())
    // Apply theme colors to message view, input field, etc.
}
```

**Key Binding Pattern** (Following k9s patterns):
```go
func (c *Chat) bindKeys() {
    c.actions.Bulk(ui.KeyMap{
        tcell.KeyEscape: ui.NewKeyAction("Close Chat", c.closeCmd, true),
        tcell.KeyEnter:  ui.NewKeyAction("Send Message", c.sendCmd, true),
        tcell.KeyCtrlC:  ui.NewKeyAction("Clear Input", c.clearCmd, true),
        tcell.KeyCtrlL:  ui.NewKeyAction("Clear Chat", c.clearChatCmd, true),
    })
}
```

#### 4. LLM Integration (Unchanged)
- **Provider Interface**: Abstraction for different LLM providers
- **OpenAI Implementation**: Primary provider with configurable API
- **Context Injection**: Current k9s state passed to LLM
- **Response Processing**: Handle various response types

#### 5. Command System (Enhanced)
- **Shell Tool**: LLM can suggest and execute shell commands
- **Approval Mechanism**: Use k9s native dialog patterns
- **Command History**: Track executed commands
- **Error Handling**: Use k9s Flash system for errors

#### 6. State Management (K9s Native)
- **K9s Context Tracker**: Monitor current namespace, resource, view
- **Chat State**: Isolated in chat component
- **Theme Integration**: Full k9s styling support

## Implementation Plan (Revised)

### Phase 1: K9s Native Foundation
**Goal**: Replace current Charmbracelet implementation with k9s native components

#### Tasks:

1. **Remove Charmbracelet Dependencies**
   ```bash
   # Remove from go.mod
   # github.com/charmbracelet/bubbletea
   # github.com/charmbracelet/bubbles  
   # github.com/charmbracelet/lipgloss
   # github.com/charmbracelet/glamour
   # github.com/charmbracelet/huh
   ```

2. **Rewrite Chat UI Components**
   
   **Create Native Chat View** (`internal/chat/ui/chat_view.go`):
   ```go
   type ChatView struct {
       *tview.Flex
       messageList *MessageList
       inputField  *InputField
       actions     *ui.KeyActions
       styles      *config.Styles
   }
   ```

   **Message Display Component** (`internal/chat/ui/message_list.go`):
   ```go
   type MessageList struct {
       *tview.TextView
       messages []Message
       styles   *config.Styles
   }
   
   func (m *MessageList) AddMessage(msg Message) {
       // Format message with k9s styling
       formatted := m.formatMessage(msg)
       // Update TextView content
       m.SetText(m.GetText(false) + formatted)
       m.ScrollToEnd()
   }
   ```

   **Input Component** (`internal/chat/ui/input_field.go`):
   ```go
   type InputField struct {
       *tview.InputField
       onSubmit func(string)
       styles   *config.Styles
   }
   ```

3. **Implement K9s Native Styling**
   
   **Add Chat Styles to Config** (`internal/config/styles.go`):
   ```go
   type Views struct {
       Table  Table  `json:"table" yaml:"table"`
       Chat   Chat   `json:"chat" yaml:"chat"`    // Add this
       // ... existing fields
   }
   
   type Chat struct {
       FgColor           Color `json:"fgColor" yaml:"fgColor"`
       BgColor           Color `json:"bgColor" yaml:"bgColor"`
       UserMessageColor  Color `json:"userMessageColor" yaml:"userMessageColor"`
       BotMessageColor   Color `json:"botMessageColor" yaml:"botMessageColor"`
       TimestampColor    Color `json:"timestampColor" yaml:"timestampColor"`
       ErrorColor        Color `json:"errorColor" yaml:"errorColor"`
   }
   ```

4. **Update Component Integration**
   
   **Rewrite ChatComponent** (`internal/chat/component.go`):
   ```go
   type ChatComponent struct {
       *tview.Flex
       chatView *ui.ChatView
       app      *view.App
       actions  *ui.KeyActions
       styles   *config.Styles
   }
   
   // Implement model.Component interface properly
   func (c *ChatComponent) Init(ctx context.Context) error {
       c.chatView = ui.NewChatView(c.app)
       c.AddItem(c.chatView, 0, 1, true)
       c.setupStyling()
       return c.chatView.Init(ctx)
   }
   ```

### Phase 2: Enhanced Features
**Goal**: Add advanced chat features using k9s patterns

#### Tasks:

1. **Command Approval System**
   - Use k9s native dialog patterns (`internal/ui/dialog/`)
   - Integrate with k9s Flash system for messages

2. **Context Awareness Enhancement**
   - Deep integration with k9s state
   - Real-time context updates

3. **Message History & Persistence**
   - Follow k9s configuration patterns
   - Store in k9s config directory

4. **Advanced LLM Features**
   - Multiple provider support
   - Streaming responses (if supported by providers)

### Phase 3: Polish & Integration
**Goal**: Perfect integration with k9s ecosystem

#### Tasks:

1. **Theme Integration**
   - Test with all k9s skins
   - Ensure accessibility compliance

2. **Performance Optimization**
   - Message virtualization for large histories
   - Efficient rendering

3. **Documentation**
   - User documentation
   - Developer documentation for extending

## Key Files to Modify/Create

### New Files (K9s Native)
```
internal/chat/
├── component.go              # Main component (rewrite)
├── ui/
│   ├── chat_view.go         # Main chat view
│   ├── message_list.go      # Message display
│   ├── input_field.go       # Input handling
│   └── styles.go            # Styling helpers
├── model/
│   ├── message.go           # Message structures
│   ├── context.go           # K9s context integration  
│   └── state.go             # State management
└── provider/                # Keep existing LLM providers
    ├── interface.go
    ├── openai.go
    └── mock.go
```

### Modified Files
```
internal/config/styles.go     # Add chat styling support
internal/view/app.go          # Layout management (already modified)
internal/view/cmd/            # Command integration (already done)
go.mod                        # Remove Charmbracelet deps
```

## Benefits of K9s Native Approach

1. **Consistency**: Matches k9s look and feel exactly
2. **Performance**: No overhead from Bubble Tea framework
3. **Maintenance**: Uses same patterns as rest of k9s
4. **Themes**: Full integration with k9s theming system
5. **Accessibility**: Inherits k9s accessibility features
6. **Size**: Smaller binary without extra dependencies
7. **Integration**: Natural fit with k9s architecture

## Styling Examples

### Chat Color Scheme Integration
```yaml
# In k9s skin files
k9s:
  views:
    chat:
      fgColor: aqua
      bgColor: black
      userMessageColor: dodgerblue
      botMessageColor: seagreen
      timestampColor: gray
      errorColor: orangered
```

### Message Formatting (tview tags)
```go
func (m *MessageList) formatUserMessage(content string) string {
    timestamp := time.Now().Format("15:04")
    return fmt.Sprintf("[%s]👤 [%s][%s::-]You[white]: %s [%s]%s[-]\n",
        m.styles.Chat().UserMessageColor,
        m.styles.Chat().UserMessageColor,
        m.styles.Chat().UserMessageColor,
        content,
        m.styles.Chat().TimestampColor,
        timestamp)
}

func (m *MessageList) formatBotMessage(content string) string {
    timestamp := time.Now().Format("15:04")
    return fmt.Sprintf("[%s]🤖 [%s][%s::-]Assistant[white]: %s [%s]%s[-]\n",
        m.styles.Chat().BotMessageColor,
        m.styles.Chat().BotMessageColor,
        m.styles.Chat().BotMessageColor,
        content,
        m.styles.Chat().TimestampColor,
        timestamp)
}
```

## Migration Strategy

1. **Keep Current Interface**: Maintain same `:chat` command and basic functionality
2. **Gradual Replacement**: Replace Charmbracelet components one by one
3. **Preserve State**: Ensure no disruption to existing chat sessions

## Future Enhancements

**Multi-Provider**: Support for multiple LLM providers simultaneously  
**Chat History**: 

This plan ensures the chat feature feels like a natural part of k9s while providing powerful AI assistance capabilities.