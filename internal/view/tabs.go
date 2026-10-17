// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const maxTabs = 9

// tab holds the per-tab state. Only the active tab's state is mirrored on App.
type tab struct {
	content       *PageStack
	cmdHistory    *model.History
	filterHistory *model.History
}

func newTab() *tab {
	return &tab{
		content:       NewPageStack(),
		cmdHistory:    model.NewHistory(model.MaxHistory),
		filterHistory: model.NewHistory(model.MaxHistory),
	}
}

// tabBar renders the tab strip. It is hidden while there is a single tab.
type tabBar struct {
	*tview.TextView

	app *App
	// spans holds each label's [start, end) columns, relative to the inner rect.
	spans [][2]int
}

func newTabBar(a *App) *tabBar {
	t := tabBar{TextView: tview.NewTextView(), app: a}
	t.SetDynamicColors(true)
	t.SetBorderPadding(0, 0, 1, 1)
	a.Styles.AddListener(&t)

	return &t
}

// StylesChanged notifies skin changed.
func (t *tabBar) StylesChanged(*config.Styles) { t.refresh() }

// StackPushed refreshes the active tab's title.
func (t *tabBar) StackPushed(model.Component) { t.refresh() }

// StackPopped refreshes the active tab's title.
func (t *tabBar) StackPopped(_, _ model.Component) { t.refresh() }

// StackTop is a noop.
func (*tabBar) StackTop(model.Component) {}

func (t *tabBar) refresh() {
	a := t.app
	st := a.Styles
	t.SetBackgroundColor(st.BgColor())
	t.Clear()
	t.spans = t.spans[:0]
	x := 0
	for i, tb := range a.tabs {
		name := "-"
		if top := tb.content.Top(); top != nil {
			name = top.Name()
		}
		bg := st.Tab().BgColor
		if i == a.curTab {
			bg = st.Tab().ActiveColor
		}
		label := fmt.Sprintf(" %d:%s ", i+1, name)
		_, _ = fmt.Fprintf(t, "[%s:%s:b]%s[-:%s:-] ", st.Tab().FgColor, bg, label, st.Body().BgColor)
		// shortcut: rune count assumes single-width names, use a width-aware count if kinds go wide.
		w := utf8.RuneCountInString(label)
		t.spans = append(t.spans, [2]int{x, x + w})
		x += w + 1
	}
}

// MouseHandler switches to the clicked tab.
func (t *tabBar) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return t.WrapMouseHandler(func(action tview.MouseAction, evt *tcell.EventMouse, _ func(tview.Primitive)) (bool, tview.Primitive) {
		x, y := evt.Position()
		if action != tview.MouseLeftClick || !t.InRect(x, y) {
			return false, nil
		}
		a := t.app
		if a.Prompt().InCmdMode() || a.Content.IsTopDialog() {
			return false, nil
		}
		ix, _, _, _ := t.GetInnerRect()
		for i, s := range t.spans {
			if x-ix >= s[0] && x-ix < s[1] {
				a.activateTab(i)
				return true, nil
			}
		}

		return false, nil
	})
}

func (a *App) syncTabBar() {
	h := 0
	if len(a.tabs) > 1 {
		h = 1
	}
	a.body.ResizeItem(a.tabBar, h, 0)
	a.tabBar.refresh()
}

// topComponent is safe to call from any goroutine.
func (a *App) topComponent() model.Component {
	a.tabMX.RLock()
	defer a.tabMX.RUnlock()

	return a.Content.Top()
}

func (a *App) detachTab() {
	if top := a.Content.Top(); top != nil {
		top.Stop()
	}
	a.Content.RemoveListener(a.Crumbs())
	a.Content.RemoveListener(a.Menu())
	a.Content.RemoveListener(a.tabBar)
}

func (a *App) attachTab(i int) {
	t := a.tabs[i]
	a.tabMX.Lock()
	a.curTab, a.Content = i, t.content
	a.tabMX.Unlock()
	a.cmdHistory, a.filterHistory = t.cmdHistory, t.filterHistory

	a.body.RemoveItemAtIndex(1)
	a.body.AddItemAtIndex(1, a.Content, 0, 1, true)
	a.Crumbs().Reset(a.Content.Peek())
	if top := a.Content.Top(); top != nil {
		a.Content.StackTop(top)
	}
	a.Content.AddListener(a.Crumbs())
	a.Content.AddListener(a.Menu())
	a.Content.AddListener(a.tabBar)
	a.syncTabBar()
}

// activateTab stops the current tab's top view and resumes the target's.
func (a *App) activateTab(i int) {
	if i < 0 || i >= len(a.tabs) || i == a.curTab {
		return
	}
	a.detachTab()
	a.attachTab(i)
}

// removeTab drops a non-active tab and releases all of its views.
func (a *App) removeTab(i int) {
	a.tabs[i].content.Close()
	a.tabs = slices.Delete(a.tabs, i, i+1)
	if i < a.curTab {
		a.curTab--
	}
	a.syncTabBar()
}

// resetTabs drops every tab but the active one. Views are bound to the
// current context so they can't survive a context switch.
func (a *App) resetTabs() {
	for i := len(a.tabs) - 1; i >= 0; i-- {
		if i != a.curTab {
			a.removeTab(i)
		}
	}
}

// confirmCloseTabs runs fn now, or once the user agrees to close the other tabs.
func (a *App) confirmCloseTabs(ctx string, fn func()) {
	if len(a.tabs) == 1 {
		fn()
		return
	}
	d := a.Styles.Dialog()
	msg := fmt.Sprintf("Switch to context %q?\nAll tabs except the current one will be closed.", ctx)
	dialog.ShowConfirm(&d, a.Content.Pages, "Switch Context", msg, fn, func() {})
}

func (a *App) addTab() error {
	if len(a.tabs) >= maxTabs {
		return fmt.Errorf("tab limit reached (%d)", maxTabs)
	}
	t := newTab()
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	if err := t.content.Init(ctx); err != nil {
		return err
	}
	a.tabs = append(a.tabs, t)
	a.activateTab(len(a.tabs) - 1)

	return nil
}

// openTab adds a tab showing the active view.
func (a *App) openTab() error {
	prev := a.curTab
	if err := a.addTab(); err != nil {
		return err
	}
	last := len(a.tabs) - 1

	err := a.command.defaultCmd(false)
	if err == nil && a.Content.Empty() {
		err = errors.New("unable to open view in new tab")
	}
	if err != nil {
		a.activateTab(prev)
		a.removeTab(last)
	}

	return err
}

func (a *App) closeTab() error {
	if len(a.tabs) == 1 {
		return errors.New("can't close the last tab")
	}
	cur := a.curTab
	next := cur + 1
	if next >= len(a.tabs) {
		next = cur - 1
	}
	a.activateTab(next)
	a.removeTab(cur)

	return nil
}

func (a *App) newTabCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.Prompt().InCmdMode() {
		return evt
	}
	if err := a.openTab(); err != nil {
		a.Flash().Err(err)
	}

	return nil
}

func (a *App) closeTabCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.Prompt().InCmdMode() {
		return evt
	}
	if err := a.closeTab(); err != nil {
		a.Flash().Err(err)
	}

	return nil
}

func (a *App) gotoTabCmd(i int) func(*tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		// Alt-<digit> shares its key code with plain runes like 'Ä', so check the modifier.
		if evt.Modifiers() != tcell.ModAlt || a.Prompt().InCmdMode() {
			return evt
		}
		if i >= len(a.tabs) {
			a.Flash().Warnf("No tab %d", i+1)
			return nil
		}
		a.activateTab(i)

		return nil
	}
}

// altKey mirrors the Alt-modified key encoding of ui.AsKey.
func altKey(r rune) tcell.Key {
	return tcell.Key(int16(r) * int16(tcell.ModAlt))
}
