# Split Panel Chat Implementation (Archive)

This document contains the split panel chat implementation logic that was replaced with the native k9s full-screen approach.

## Original Split Panel Logic

### buildContentArea() - Split Panel Implementation
```go
func (a *App) buildContentArea() tview.Primitive {
	if a.chatVisible && a.chatComponent != nil {
		// Split layout: main content on left, chat on right
		splitPane := tview.NewFlex().SetDirection(tview.FlexColumn)
		splitPane.AddItem(a.Content, 0, 7, !a.chatFocused)      // 70% for main content
		splitPane.AddItem(a.chatComponent, 0, 3, a.chatFocused) // 30% for chat

		// Set up Tab key handling for focus switching
		splitPane.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				a.toggleChatFocus()
				return nil
			}

			// If chat is focused, prevent k9s shortcuts from being processed
			if a.chatFocused {
				// Only allow Escape to close chat, Tab to switch focus
				if event.Key() == tcell.KeyEscape {
					a.chatCmd() // Close chat
					return nil
				}
				// Forward all other events to chat component
				if a.chatComponent != nil {
					// Let the chat component handle the event
					return event
				}
				return nil // Consume the event to prevent k9s processing
			}

			return event // Let k9s handle the event normally
		})

		return splitPane
	}
	// Normal layout: just the main content
	return a.Content
}
```

### Original Focus Management Logic
```go
func (a *App) toggleChatFocus() {
	// Check if chat is on top of stack (focused)
	if top := a.Content.Top(); top != nil && top.Name() == "chat" {
		// Chat is focused, switch to main k9s view
		a.Content.Pop() // Remove chat from stack
		a.chatFocused = false
		a.QueueUpdateDraw(func() {
			a.SetFocus(a.Content) // Focus main content
		})
	} else if a.chatVisible && a.chatComponent != nil {
		// Chat is visible but not focused, focus it
		if err := a.inject(a.chatComponent, false); err != nil {
			slog.Error("Failed to focus chat", slogs.Error, err)
			return
		}
		a.chatFocused = true
		a.QueueUpdateDraw(func() {
			a.SetFocus(a.chatComponent) // Focus chat
		})
	}
}
```

## Benefits of Split Panel Approach
- Both k9s and chat visible simultaneously  
- Quick context switching with Tab key
- Maintains k9s workflow while chatting

## Issues with Split Panel Approach
- Complex focus management with dual components
- Layout rebuilding complexity
- Not following pure k9s native patterns
- Screen space constraints on smaller terminals
- Inconsistent with other k9s views (Help, Alias, etc.)

## Reasons for Migration to Full-Screen
1. **Native k9s Pattern**: Full-screen views like Help, Alias, Browser
2. **Simplicity**: Single component focus, no complex layout management  
3. **Screen Real Estate**: Better space utilization for chat content
4. **Consistency**: Same behavior as other k9s overlay views
5. **Maintainability**: Leverages existing k9s infrastructure

## Future Considerations
The split panel approach could be revisited as an optional mode or advanced feature if there's user demand for simultaneous k9s/chat viewing.
