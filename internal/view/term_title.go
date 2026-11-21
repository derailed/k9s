// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"strings"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui/termtitle"
)

const termTitlePrefix = "k9s"

// termTitle keeps the terminal title in sync with the active context and view
// when ui.terminalTitle is enabled.
type termTitle struct {
	app   *App
	title *termtitle.Title
}

func newTermTitle(a *App) *termTitle {
	return &termTitle{app: a, title: termtitle.New()}
}

// StackPushed indicates a new item was added.
func (t *termTitle) StackPushed(c model.Component) {
	t.update(c)
}

// StackPopped indicates an item was deleted.
func (t *termTitle) StackPopped(_, top model.Component) {
	t.update(top)
}

// StackTop indicates the top of the stack.
func (*termTitle) StackTop(model.Component) {}

func (t *termTitle) update(top model.Component) {
	if !t.app.Config.K9s.UI.TerminalTitle {
		t.title.Restore()
		return
	}
	var view string
	if top != nil {
		view = top.Name()
	}
	t.title.Set(formatTermTitle(t.app.Config.ActiveContextName(), view))
}

func (t *termTitle) restore() {
	if t == nil {
		return
	}
	t.title.Restore()
}

// formatTermTitle builds "k9s | <context> | <view>", skipping empty parts.
// The view name is normalized the same way the crumbs show it.
func formatTermTitle(ctx, view string) string {
	parts := []string{termTitlePrefix}
	if ctx != "" {
		parts = append(parts, ctx)
	}
	if view = strings.ToLower(strings.ReplaceAll(view, " ", "")); view != "" {
		parts = append(parts, view)
	}

	return strings.Join(parts, " | ")
}
