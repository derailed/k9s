// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const maxTabs = 9

// TabManager orchestrates multiple tab sessions within a shared tview layout.
//
// Layout hierarchy (replaces the bare Content slot in the main Flex):
//
//	contentArea  *tview.Flex  (FlexRow)
//	  ├── tabBar  *ui.TabBar    (1 line, visible only when len(sessions) > 1)
//	  └── container *tview.Pages (one page per TabSession)
//
// The K8s factory, cluster model, and application styles are shared across all
// sessions.  Each session owns its view stack, command interpreter and
// navigation histories.
//
// Concurrency: every method on TabManager must be called from the tview main
// goroutine.  Call sites running on other goroutines must wrap invocations in
// app.QueueUpdateDraw.  This single-writer contract is what allows the type
// to omit explicit synchronization.
type TabManager struct {
	sessions      []*TabSession
	activeIdx     int
	nextID        int
	container     *tview.Pages
	tabBar        *ui.TabBar
	contentArea   *tview.Flex
	tabBarVisible bool
	app           *App
}

// newTabManager constructs a TabManager without any sessions.
// Call initWithSession to seed the first session.
func newTabManager(app *App) *TabManager {
	tm := &TabManager{
		app:       app,
		container: tview.NewPages(),
		tabBar:    ui.NewTabBar(app.Styles),
	}
	tm.contentArea = tview.NewFlex().SetDirection(tview.FlexRow)
	tm.contentArea.AddItem(tm.container, 0, 1, true)
	return tm
}

// initWithSession registers sess as the first tab.  The session's Content,
// command and histories must already be initialised by the caller (app.Init
// does this before calling initWithSession).  Crumbs and Menu listeners have
// already been attached to sess.Content by app.Init as well.
func (tm *TabManager) initWithSession(sess *TabSession) {
	sess.id = tm.nextID
	tm.nextID++
	tm.sessions = []*TabSession{sess}
	tm.activeIdx = 0
	tm.container.AddPage(tm.pageKey(sess.id), sess.Content, true, true)
}

// Active returns the currently active session.
func (tm *TabManager) Active() *TabSession {
	return tm.sessions[tm.activeIdx]
}

// Count returns the number of open tabs.
func (tm *TabManager) Count() int {
	return len(tm.sessions)
}

// newTab creates a new session, activates it, and registers its page in the
// container.  The caller is responsible for navigating the new session to the
// desired resource after newTab returns.
// Must be called on the tview main goroutine.
func (tm *TabManager) newTab(ctx context.Context) error {
	if len(tm.sessions) >= maxTabs {
		return fmt.Errorf("maximum %d tabs allowed", maxTabs)
	}

	sess := &TabSession{
		id:            tm.nextID,
		Content:       NewPageStack(),
		cmdHistory:    model.NewHistory(model.MaxHistory),
		filterHistory: model.NewHistory(model.MaxHistory),
		label:         "",
	}

	if err := sess.Content.Init(ctx); err != nil {
		return fmt.Errorf("init tab content: %w", err)
	}

	cmd := NewCommand(tm.app)
	if err := cmd.Init(tm.app.Config.ContextAliasesPath()); err != nil {
		// Roll back the already-initialised PageStack so it doesn't outlive
		// this failed newTab call as an orphaned, listener-attached primitive.
		sess.Content.Clear()
		return fmt.Errorf("init tab command: %w", err)
	}
	sess.command = cmd

	tm.nextID++
	tm.sessions = append(tm.sessions, sess)
	newIdx := len(tm.sessions) - 1

	tm.container.AddPage(tm.pageKey(sess.id), sess.Content, true, false)
	tm.activateSession(newIdx)
	tm.refreshTabBar()

	return nil
}

// closeActive closes the active tab, switches to an adjacent one, and stops
// all components belonging to the closed session.
// Returns an error when the last tab is requested to be closed.
// Must be called on the tview main goroutine.
func (tm *TabManager) closeActive() error {
	if len(tm.sessions) <= 1 {
		return fmt.Errorf("cannot close the last tab")
	}

	idx := tm.activeIdx
	sess := tm.sessions[idx]

	// Prefer the right neighbour; fall back to the left one.
	nextIdx := idx
	if idx >= len(tm.sessions)-1 {
		nextIdx = idx - 1
	}
	// After removal, an element originally to the right of idx shifts left by
	// one, landing at idx — so nextIdx stays correct when nextIdx == idx.

	tm.sessions = append(tm.sessions[:idx], tm.sessions[idx+1:]...)

	// Switch app state to the target session (rewires listeners).
	tm.activateSession(nextIdx)

	// Stop all view components belonging to the closed session.
	// Clear() internally calls StackTop() which may redirect focus to
	// intermediate components of the now-closed session.  Restore focus
	// explicitly afterwards.
	sess.Content.Clear()
	tm.container.RemovePage(tm.pageKey(sess.id))

	// Re-establish focus on the new session's top component, which may have
	// been overwritten by the PageStack.StackTop callbacks fired during Clear().
	if top := tm.app.Content.Top(); top != nil {
		tm.app.SetFocus(top)
	}

	tm.refreshTabBar()
	return nil
}

// CloseOtherTabs closes all tabs except the currently active one.
// Must be called on the tview main goroutine.
func (tm *TabManager) CloseOtherTabs() {
	if len(tm.sessions) <= 1 {
		return
	}

	activeSess := tm.sessions[tm.activeIdx]
	toClose := make([]*TabSession, 0, len(tm.sessions)-1)
	for i, sess := range tm.sessions {
		if i != tm.activeIdx {
			toClose = append(toClose, sess)
		}
	}

	tm.sessions = []*TabSession{activeSess}
	tm.activeIdx = 0

	for _, sess := range toClose {
		// Defensively detach app-wide listeners before Clear(), since Clear
		// fires StackPopped → StackTop callbacks on every remaining listener.
		// The invariant today is that only the active tab carries Crumbs/Menu
		// listeners (activateSession hands them over on every switch), so
		// these RemoveListener calls are normally no-ops, but they protect us
		// from regressions if that invariant ever loosens.
		sess.Content.RemoveListener(tm.app.Crumbs())
		sess.Content.RemoveListener(tm.app.Menu())
		sess.Content.Clear()
		tm.container.RemovePage(tm.pageKey(sess.id))
	}

	// Re-establish focus on the new session's top component, which may have
	// been overwritten by the PageStack.StackTop callbacks fired during Clear().
	if top := tm.app.Content.Top(); top != nil {
		tm.app.SetFocus(top)
	}

	tm.refreshTabBar()
}

// SwitchTo activates the tab at the given zero-based slice index.
// Must be called on the tview main goroutine.
func (tm *TabManager) SwitchTo(idx int) {
	if idx < 0 || idx >= len(tm.sessions) {
		return
	}
	tm.activateSession(idx)
	tm.refreshTabBar()
}

// NextTab activates the tab to the right, wrapping around.
func (tm *TabManager) NextTab() {
	count := len(tm.sessions)
	if count <= 1 {
		return
	}
	tm.SwitchTo((tm.activeIdx + 1) % count)
}

// PrevTab activates the tab to the left, wrapping around.
func (tm *TabManager) PrevTab() {
	count := len(tm.sessions)
	if count <= 1 {
		return
	}
	tm.SwitchTo((tm.activeIdx - 1 + count) % count)
}

// updateActiveLabel updates the label shown for the active tab.
// Must be called on the tview main goroutine.
func (tm *TabManager) updateActiveLabel(label string) {
	tm.sessions[tm.activeIdx].label = label
	tm.refreshTabBar()
}

// activateSession rewires the app's mutable state pointers (Content, command,
// histories) to the session at idx and brings its page to the front.
// Must be called on the tview main goroutine.
func (tm *TabManager) activateSession(idx int) {
	if idx < 0 || idx >= len(tm.sessions) {
		return
	}

	app := tm.app
	newSess := tm.sessions[idx]
	oldContent := app.Content

	// Guard: if the target session is already active there is nothing to do.
	if oldContent == newSess.Content {
		tm.activeIdx = idx
		return
	}
	tm.activeIdx = idx

	// Stop the outgoing tab's top component so its data-watch goroutines,
	// informers and tickers terminate.  Without this every tab switch would
	// leak goroutines and stack watchers on the API server.
	if oldContent != nil {
		if top := oldContent.Top(); top != nil {
			top.Stop()
		}
	}

	// Detach breadcrumbs and menu from the outgoing content so they no longer
	// receive push/pop events from the tab we are leaving.
	if oldContent != nil {
		oldContent.RemoveListener(app.Crumbs())
		oldContent.RemoveListener(app.Menu())
	}

	// Swap the active-session pointers in App.  All existing code paths that
	// reference app.Content / app.command / app.cmdHistory / app.filterHistory
	// automatically operate on the new tab from this point on.
	app.Content = newSess.Content
	app.command = newSess.command
	app.cmdHistory = newSess.cmdHistory
	app.filterHistory = newSess.filterHistory

	// Rebuild breadcrumbs to reflect the incoming tab's navigation history.
	app.Crumbs().Reset(newSess.Content.Peek())

	// Attach breadcrumbs and menu to the incoming content.
	newSess.Content.AddListener(app.Crumbs())
	newSess.Content.AddListener(app.Menu())

	// Bring the new session's page to the front of the container.
	tm.container.SwitchToPage(tm.pageKey(newSess.id))

	// Restart and focus the top-most component of the incoming session.
	if top := newSess.Content.Top(); top != nil {
		top.Start()
		app.SetFocus(top)
	}
}

// refreshTabBar shows or hides the tab bar strip and re-renders its labels.
// Must be called on the tview main goroutine.
func (tm *TabManager) refreshTabBar() {
	count := len(tm.sessions)
	labels := make([]string, count)
	for i, s := range tm.sessions {
		labels[i] = s.label
	}

	showBar := count > 1
	switch {
	case showBar && !tm.tabBarVisible:
		tm.contentArea.AddItemAtIndex(0, tm.tabBar, 1, 1, false)
		tm.tabBarVisible = true
	case !showBar && tm.tabBarVisible:
		tm.contentArea.RemoveItemAtIndex(0)
		tm.tabBarVisible = false
	}
	if showBar {
		tm.tabBar.Render(labels, tm.activeIdx)
	}
}

func (tm *TabManager) pageKey(id int) string {
	return fmt.Sprintf("tab-%d", id)
}

// switchNS switches the namespace for all open sessions.
// Must be called on the tview main goroutine.
func (tm *TabManager) switchNS(ns string) {
	for _, sess := range tm.sessions {
		sess.cmdHistory.SwitchNS(ns)
		for _, c := range sess.Content.Peek() {
			rv, ok := c.(ResourceViewer)
			if !ok {
				continue
			}
			if namespaced, err := dao.MetaAccess.IsNamespaced(rv.GVR()); err == nil && !namespaced {
				continue
			}
			if b, ok := rv.(*Browser); ok {
				b.setNamespace(ns)
				continue
			}
			tbl := rv.GetTable()
			if tbl == nil {
				continue
			}
			m := tbl.GetModel()
			if m == nil {
				continue
			}
			m.SetNamespace(ns)
		}
	}
}
