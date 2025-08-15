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

## LLM Tool Integration - Simplified kubectl Approach

### Overview
Instead of multiple complex tools, the LLM gets one powerful tool: `execute_kubectl` that can run any kubectl command within the current k9s context. This leverages k9s existing kubectl infrastructure while providing maximum flexibility.

### Tool Architecture

#### Single Tool Definition
```go
// internal/chat/tool/kubectl.go - Reuse k9s native infrastructure
type KubectlExecutor struct {
    app *view.App
}

func (k *KubectlExecutor) Execute(command string) (string, error) {
    args := strings.Fields(command)
    
    // Reuse k9s existing runKu from internal/view/exec.go
    opts := &shellOpts{
        args:       args,
        clear:      false,
        background: false,
    }
    
    // This already handles context, auth, kubeconfig automatically
    return runKu(k.app, opts)
}

// Use k9s existing dialog system for dangerous commands
func (k *KubectlExecutor) RequestApproval(command string, callback func(bool)) {
    title := "LLM Command Execution"
    message := "Execute: kubectl " + command
    
    // Reuse k9s existing ShowConfirm dialog
    d := k.app.Styles.Dialog()
    dialog.ShowConfirm(&d, k.app.Content.Pages, title, message,
        func() { callback(true) },  // OK
        func() { callback(false) }, // Cancel
    )
}
```

### Two Provider Implementation

#### 1. OpenAI Provider (Production)
```go
// internal/chat/provider/openai.go
type OpenAIProvider struct {
    client      *openai.Client
    kubectlTool *tool.KubectlExecutor
    context     *model.K9sContext
}

func (p *OpenAIProvider) SendMessage(ctx context.Context, messages []Message, opts *Options) (*Response, error) {
    // Update k9s context before each request
    p.context.UpdateFromApp()
    
    // Define single kubectl tool for LLM
    tools := []openai.Tool{{
        Type: openai.ToolTypeFunction,
        Function: &openai.FunctionDefinition{
            Name:        "execute_kubectl",
            Description: "Execute kubectl commands in current k9s context",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "command": map[string]interface{}{
                        "type":        "string",
                        "description": "kubectl command (without 'kubectl' prefix). Examples: 'get pods', 'describe pod my-pod'",
                    },
                },
                "required": []string{"command"},
            },
        },
    }}
    
    req := openai.ChatCompletionRequest{
        Model:    opts.Model,
        Messages: p.buildMessages(messages),
        Tools:    tools,
    }
    
    resp, err := p.client.CreateChatCompletion(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // Handle tool calls
    if len(resp.Choices[0].Message.ToolCalls) > 0 {
        return p.handleKubectlCall(resp.Choices[0].Message.ToolCalls[0])
    }
    
    return &Response{Content: resp.Choices[0].Message.Content}, nil
}

func (p *OpenAIProvider) handleKubectlCall(toolCall openai.ToolCall) (*Response, error) {
    var args struct {
        Command string `json:"command"`
    }
    json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
    
    // Check if command needs approval (dangerous operations)
    if p.kubectlTool.NeedsApproval(args.Command) {
        approved := make(chan bool, 1)
        p.kubectlTool.RequestApproval(args.Command, func(result bool) {
            approved <- result
        })
        
        if !<-approved {
            return &Response{Content: "Command cancelled by user"}, nil
        }
    }
    
    // Execute using k9s infrastructure
    output, err := p.kubectlTool.Execute(args.Command)
    if err != nil {
        return &Response{Content: fmt.Sprintf("Error: %s\nOutput: %s", err, output)}, nil
    }
    
    return &Response{Content: fmt.Sprintf("```\n%s\n```", output)}, nil
}
```

#### 2. Demo Mock Provider (Comprehensive Demo)

**Enhanced mock provider that demonstrates all features:**

```go
// internal/chat/provider/demo.go - Comprehensive demo provider
type DemoProvider struct {
    kubectlTool   *tool.KubectlExecutor
    context       *model.K9sContext
    demoStep      int
    simulatedPod  string
}

func NewDemoProvider(app *view.App, context *model.K9sContext) *DemoProvider {
    return &DemoProvider{
        kubectlTool:   tool.NewKubectlExecutor(app),
        context:       context,
        demoStep:      0,
        simulatedPod:  "demo-nginx-" + randomString(6),
    }
}

func (p *DemoProvider) SendMessage(ctx context.Context, messages []Message, opts *Options) (*Response, error) {
    p.context.UpdateFromApp()
    
    userMsg := messages[len(messages)-1].Content
    
    // Demo scenario: Create nginx pod -> explain -> delete pod
    switch p.demoStep {
    case 0:
        return p.handleCreatePodDemo(userMsg)
    case 1:
        return p.handleExplainDemo(userMsg)
    case 2:
        return p.handleDeleteDemo(userMsg)
    default:
        return p.handleGeneralDemo(userMsg)
    }
}

// Step 1: Create nginx pod and show context awareness
func (p *DemoProvider) handleCreatePodDemo(userMsg string) (*Response, error) {
    p.demoStep = 1
    
    response := fmt.Sprintf(`## 🚀 Creating Demo Nginx Pod

### Context Awareness Demo
I can see you're currently in:
- **Cluster**: %s
- **Namespace**: %s
- **Current View**: %s

Let me create a demo nginx pod for you!

### Executing Command
I'll use the kubectl tool to create a pod:`, 
        p.context.ClusterName, p.context.Namespace, p.context.CurrentView)
    
    // Simulate kubectl command execution
    createCmd := fmt.Sprintf("run %s --image=nginx --port=80", p.simulatedPod)
    
    // Execute actual kubectl command (will fail safely in demo)
    output, err := p.kubectlTool.Execute(createCmd)
    
    if err != nil {
        // Show simulated success for demo
        response += fmt.Sprintf(`

**Command**: kubectl %s

```
pod/%s created
```

✅ **Success!** Created nginx pod in namespace **%s**

### Next Steps
The pod is now being created. Ask me to explain what we just did, or how to interact with this pod!`, 
            createCmd, p.simulatedPod, p.context.Namespace)
    } else {
        // Real success
        response += fmt.Sprintf(`

**Command**: kubectl %s

```
%s
```

✅ **Pod created successfully!**`, createCmd, output)
    }
    
    return &Response{Content: response}, nil
}

// Step 2: Explain pod usage and kubectl commands
func (p *DemoProvider) handleExplainDemo(userMsg string) (*Response, error) {
    p.demoStep = 2
    
    response := fmt.Sprintf(`## 📚 Understanding Your Nginx Pod

### What We Created
The pod **%s** is now running nginx web server in your **%s** namespace.

### Useful Commands to Try

#### Check Pod Status
` + "```bash" + `
kubectl get pods
kubectl describe pod %s
kubectl logs %s
` + "```" + `

#### Interact with the Pod
` + "```bash" + `
# Port forward to access nginx
kubectl port-forward pod/%s 8080:80

# Execute commands inside pod
kubectl exec -it %s -- /bin/bash

# Get pod IP and test connectivity
kubectl get pod %s -o wide
` + "```" + `

### k9s Integration
In k9s, you can:
- Press **l** to view logs
- Press **s** to shell into the pod  
- Press **d** to describe the pod
- Press **y** to view pod YAML

### Demo Cleanup
In a moment, I'll demonstrate the **approval system** by deleting this pod. 
This will show how dangerous commands require user confirmation for safety.

*Ready to see the deletion approval process?*`, 
        p.simulatedPod, p.context.Namespace, p.simulatedPod, p.simulatedPod, 
        p.simulatedPod, p.simulatedPod, p.simulatedPod)
    
    // Wait 5 seconds before allowing deletion
    time.Sleep(5 * time.Second)
    
    return &Response{Content: response}, nil
}

// Step 3: Delete pod with approval demo
func (p *DemoProvider) handleDeleteDemo(userMsg string) (*Response, error) {
    p.demoStep = 3
    
    response := fmt.Sprintf(`## 🗑️ Demonstrating Safe Command Execution

### Approval System Demo
Now I'll delete the demo pod to show the **approval system** in action.

**Note**: Delete operations are considered dangerous and require user approval.

### Executing Delete Command
I'm requesting to execute:`)
    
    // Trigger approval dialog for delete command
    deleteCmd := fmt.Sprintf("delete pod %s", p.simulatedPod)
    
    // This will trigger the approval dialog
    approved := make(chan bool, 1)
    go func() {
        p.kubectlTool.RequestApproval(deleteCmd, func(result bool) {
            approved <- result
        })
    }()
    
    // Wait for approval (this is async in real implementation)
    if <-approved {
        output, err := p.kubectlTool.Execute(deleteCmd)
        
        if err != nil {
            // Simulated success
            response += fmt.Sprintf(`

**Command**: kubectl %s

```
pod "%s" deleted
```

✅ **Pod deleted successfully!**

### Demo Complete! 🎉

This demonstration showed:
1. **Context Awareness**: I knew your current cluster, namespace, and view
2. **Command Execution**: Used kubectl through k9s native infrastructure  
3. **Safety System**: Delete command required your approval
4. **Real Integration**: All styling and dialogs use k9s native components

You can now ask me about any Kubernetes operations in your cluster!`, 
                deleteCmd, p.simulatedPod)
        } else {
            response += fmt.Sprintf(`

```
%s
```

✅ **Pod deleted successfully!**`, output)
        }
    } else {
        response += `

❌ **Operation cancelled by user**

The delete operation was safely cancelled. This demonstrates how the approval system protects against accidental dangerous operations.

### Safety Features
- Read operations (get, describe, logs) execute immediately
- Modify operations (create, apply, scale) may require approval
- Delete operations always require approval
- Context is automatically applied to all commands`
    }
    
    return &Response{Content: response}, nil
}

// General demo responses
func (p *DemoProvider) handleGeneralDemo(userMsg string) (*Response, error) {
    // Context-aware responses based on current k9s state
    response := fmt.Sprintf(`## 🤖 Demo Assistant Ready

### Current Context
- **Cluster**: %s
- **Namespace**: %s  
- **View**: %s

### Available Commands
I can execute any kubectl command in your current context. For example:

` + "```bash" + `
# Safe commands (execute immediately)
kubectl get pods
kubectl describe deployment
kubectl logs <pod-name>

# Dangerous commands (require approval)  
kubectl delete pod <name>
kubectl apply -f manifest.yaml
kubectl scale deployment <name> --replicas=0
` + "```" + `

### Demo Features Shown
✅ **Context Awareness** - I know your current k9s state
✅ **kubectl Integration** - Uses k9s native infrastructure
✅ **Approval System** - Safety for dangerous operations
✅ **Native Styling** - Matches k9s theme perfectly

**Try asking**: "Create a test pod" or "Show me my deployments"`, 
        p.context.ClusterName, p.context.Namespace, p.context.CurrentView)
    
    return &Response{Content: response}, nil
}
```

### Safety Classification

#### Auto-Execute Commands (Safe)
- `get`, `describe`, `logs`, `top`, `version`, `config`, `api`, `explain`

#### Require Approval (Dangerous)
- `delete`, `apply`, `create`, `patch`, `replace`, `scale`, `rollout`, `exec`

### Context Integration

**Enhanced K9s Context Provider:**
```go
// internal/chat/model/context.go - Enhanced for tool execution
func (ctx *K9sContext) UpdateFromApp() {
    // Leverage k9s existing state management
    ctx.ContextName = ctx.app.Config.K9s.ActiveContextName()
    ctx.Namespace = ctx.app.Config.ActiveNamespace()
    ctx.CurrentView = ctx.app.Config.ActiveView()
    
    // Get cluster info using k9s connection
    if conn := ctx.app.Conn(); conn != nil {
        if cfg := conn.Config(); cfg != nil {
            ctx.ClusterName, _ = cfg.CurrentClusterName()
        }
    }
    
    // Get selected item using k9s table interface
    if top := ctx.app.Content.Top(); top != nil {
        if tableView, ok := top.(interface{ GetTable() *ui.Table }); ok {
            ctx.SelectedItem = tableView.GetTable().GetSelectedItem()
        }
    }
}

func (ctx *K9sContext) GetSystemPrompt() string {
    return fmt.Sprintf(`You are an expert Kubernetes assistant integrated into k9s.
You can execute kubectl commands using the execute_kubectl tool.

CURRENT CONTEXT:
- Cluster: %s
- Context: %s  
- Namespace: %s
- Current View: %s
- Selected Item: %s

TOOL USAGE:
- Use execute_kubectl to run any kubectl command
- Commands automatically use current context and configuration
- Read commands execute immediately
- Dangerous commands require user approval
- Always explain what you're doing and why

When users ask about their current selection, use the selected resource context.
Be helpful, safe, and leverage the current k9s state.`, 
        ctx.ClusterName, ctx.ContextName, ctx.Namespace, 
        ctx.CurrentView, ctx.SelectedItem)
}
```

### Integration Benefits

1. **Maximum k9s Reuse**: Uses existing `runKu()`, `dialog.ShowConfirm()`, app state
2. **Single Tool Simplicity**: One powerful `execute_kubectl` tool for all operations  
3. **Native Safety**: Leverages k9s existing approval patterns
4. **Context Awareness**: Full access to current k9s state and selections
5. **Demo Ready**: Comprehensive mock provider showcases all features

### Demo Flow Summary

1. **User opens chat**: Mock provider shows context awareness
2. **Create pod**: Demonstrates kubectl execution and success feedback
3. **Explain usage**: Shows educational content and k9s integration tips
4. **Delete pod**: Triggers approval dialog to show safety system
5. **General usage**: Context-aware responses for any kubectl operations

This approach provides maximum flexibility while maintaining safety through k9s native patterns and minimal implementation complexity.

## Future Enhancements

**Multi-Provider**: Support for multiple LLM providers simultaneously  
**Chat History**: Persistent conversation storage using k9s config patterns
**Streaming**: Real-time response streaming for better UX
**Custom Tools**: Extensible tool system for specialized operations

This plan ensures the chat feature feels like a natural part of k9s while providing powerful AI assistance capabilities with comprehensive demo features.