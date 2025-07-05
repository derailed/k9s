// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/k9s/internal/vul"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// ExitStatus indicates UI exit conditions.
var ExitStatus = ""

const (
	splashDelay      = 1 * time.Second
	clusterRefresh   = 15 * time.Second
	clusterInfoWidth = 50
	clusterInfoPad   = 15
)

// App represents an application view.
type App struct {
	version string
	*ui.App
	Content       *PageStack
	command       *Command
	factory       *watch.Factory
	cancelFn      context.CancelFunc
	clusterModel  *model.ClusterInfo
	cmdHistory    *model.History
	filterHistory *model.History
	conRetry      int32
	showHeader    bool
	showLogo      bool
	showCrumbs    bool
}

// NewApp returns a K9s app instance.
func NewApp(cfg *config.Config) *App {
	a := App{
		App:           ui.NewApp(cfg, cfg.K9s.ActiveContextName()),
		cmdHistory:    model.NewHistory(model.MaxHistory),
		filterHistory: model.NewHistory(model.MaxHistory),
		Content:       NewPageStack(),
	}
	a.ReloadStyles()

	a.Views()["statusIndicator"] = ui.NewStatusIndicator(a.App, a.Styles)
	a.Views()["clusterInfo"] = NewClusterInfo(&a)

	return &a
}

// ReloadStyles reloads skin file.
func (a *App) ReloadStyles() {
	a.RefreshStyles(a)
}

// UpdateClusterInfo updates clusterInfo panel
func (a *App) UpdateClusterInfo() {
	if a.factory != nil {
		a.clusterModel.Reset(a.factory)
	}
}

// ConOK checks the connection is cool, returns false otherwise.
func (a *App) ConOK() bool {
	return atomic.LoadInt32(&a.conRetry) == 0
}

// Init initializes the application.
func (a *App) Init(version string, _ int) error {
	a.version = model.NormalizeVersion(version)

	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	if err := a.Content.Init(ctx); err != nil {
		return err
	}
	a.Content.AddListener(a.Crumbs())
	a.Content.AddListener(a.Menu())

	a.App.Init()
	a.SetInputCapture(a.keyboard)
	a.bindKeys()
	if a.Conn() == nil {
		return errors.New("no client connection detected")
	}
	ns := a.Config.ActiveNamespace()

	a.factory = watch.NewFactory(a.Conn())
	a.initFactory(ns)

	a.clusterModel = model.NewClusterInfo(a.factory, a.version, a.Config.K9s)
	a.clusterModel.AddListener(a.clusterInfo())
	a.clusterModel.AddListener(a.statusIndicator())
	if a.Conn().ConnectionOK() {
		a.clusterModel.Refresh()
		a.clusterInfo().Init()
	}

	a.command = NewCommand(a)
	if err := a.command.Init(a.Config.ContextAliasesPath()); err != nil {
		return err
	}
	a.CmdBuff().SetSuggestionFn(a.suggestCommand())

	a.layout(ctx)
	a.initSignals()

	if a.Config.K9s.ImageScans.Enable {
		a.initImgScanner(version)
	}
	a.ReloadStyles()

	return nil
}

func (*App) stopImgScanner() {
	if vul.ImgScanner != nil {
		vul.ImgScanner.Stop()
	}
}

func (a *App) initImgScanner(version string) {
	defer func(t time.Time) {
		slog.Debug("Scanner init time", slogs.Elapsed, time.Since(t))
	}(time.Now())

	vul.ImgScanner = vul.NewImageScanner(a.Config.K9s.ImageScans, slog.Default())
	go vul.ImgScanner.Init("k9s", version)
}

func (a *App) layout(ctx context.Context) {
	flash := ui.NewFlash(a.App)
	go flash.Watch(ctx, a.Flash().Channel())

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(a.statusIndicator(), 1, 1, false)
	main.AddItem(a.Content, 0, 10, true)
	if !a.Config.K9s.IsCrumbsless() {
		main.AddItem(a.Crumbs(), 1, 1, false)
	}
	main.AddItem(flash, 1, 1, false)

	a.Main.AddPage("main", main, true, false)
	a.toggleHeader(!a.Config.K9s.IsHeadless(), !a.Config.K9s.IsLogoless())
	if !a.Config.K9s.IsSplashless() {
		a.Main.AddPage("splash", ui.NewSplash(a.Styles, a.version), true, true)
	}
}

func (*App) initSignals() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)

	go func(sig chan os.Signal) {
		<-sig
		os.Exit(0)
	}(sig)
}

func (a *App) suggestCommand() model.SuggestionFunc {
	contextNames, err := a.contextNames()
	if err != nil {
		slog.Error("Failed to list contexts", slogs.Error, err)
	}

	return func(s string) (entries sort.StringSlice) {
		if s == "" {
			if a.cmdHistory.Empty() {
				return
			}
			return a.cmdHistory.List()
		}

		ls := strings.ToLower(s)
		for alias := range maps.Keys(a.command.alias.Alias) {
			if suggest, ok := cmd.ShouldAddSuggest(ls, alias); ok {
				entries = append(entries, suggest)
			}
		}

		namespaceNames, err := a.factory.Client().ValidNamespaceNames()
		if err != nil {
			slog.Error("Failed to obtain list of namespaces", slogs.Error, err)
		}
		entries = append(entries, cmd.SuggestSubCommand(s, namespaceNames, contextNames)...)
		if len(entries) == 0 {
			return nil
		}
		entries.Sort()
		return
	}
}

func (a *App) contextNames() ([]string, error) {
	contexts, err := a.factory.Client().Config().Contexts()
	if err != nil {
		return nil, err
	}
	contextNames := make([]string, 0, len(contexts))
	for ctxName := range contexts {
		contextNames = append(contextNames, ctxName)
	}

	return contextNames, nil
}

func (a *App) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if k, ok := a.HasAction(ui.AsKey(evt)); ok && !a.Content.IsTopDialog() {
		return k.Action(evt)
	}

	return evt
}

func (a *App) bindKeys() {
	a.AddActions(ui.NewKeyActionsFromMap(ui.KeyMap{
		ui.KeyShift9:       ui.NewSharedKeyAction("DumpGOR", a.dumpGOR, false),
		tcell.KeyCtrlE:     ui.NewSharedKeyAction("ToggleHeader", a.toggleHeaderCmd, false),
		tcell.KeyCtrlG:     ui.NewSharedKeyAction("toggleCrumbs", a.toggleCrumbsCmd, false),
		ui.KeyHelp:         ui.NewSharedKeyAction("Help", a.helpCmd, false),
		ui.KeyLeftBracket:  ui.NewSharedKeyAction("Go Back", a.previousCommand, false),
		ui.KeyRightBracket: ui.NewSharedKeyAction("Go Forward", a.nextCommand, false),
		ui.KeyDash:         ui.NewSharedKeyAction("Last View", a.lastCommand, false),
		tcell.KeyCtrlA:     ui.NewSharedKeyAction("Aliases", a.aliasCmd, false),
		tcell.KeyEnter:     ui.NewKeyAction("Goto", a.gotoCmd, false),
		tcell.KeyCtrlC:     ui.NewKeyAction("Quit", a.quitCmd, false),
	}))
}

func (*App) dumpGOR(evt *tcell.EventKey) *tcell.EventKey {
	slog.Debug("GOR", slogs.GOR, runtime.NumGoroutine())
	bb := make([]byte, 5_000_000)
	runtime.Stack(bb, true)
	slog.Debug("GOR stack", slogs.Stack, string(bb))

	return evt
}

// ActiveView returns the currently active view.
func (a *App) ActiveView() model.Component {
	return a.Content.GetPrimitive("main").(model.Component)
}

func (a *App) toggleHeader(header, logo bool) {
	a.showHeader, a.showLogo = header, logo
	flex, ok := a.Main.GetPrimitive("main").(*tview.Flex)
	if !ok {
		slog.Error("Expecting flex view main panel. Exiting!")
		os.Exit(1)
	}
	if a.showHeader {
		flex.RemoveItemAtIndex(0)
		flex.AddItemAtIndex(0, a.buildHeader(), 7, 1, false)
	} else {
		flex.RemoveItemAtIndex(0)
		flex.AddItemAtIndex(0, a.statusIndicator(), 1, 1, false)
	}
}

func (a *App) toggleCrumbs(flag bool) {
	a.showCrumbs = flag
	flex, ok := a.Main.GetPrimitive("main").(*tview.Flex)
	if !ok {
		slog.Error("Expecting valid flex view main panel. Exiting!")
		os.Exit(1)
	}
	if a.showCrumbs {
		if _, ok := flex.ItemAt(2).(*ui.Crumbs); !ok {
			flex.AddItemAtIndex(2, a.Crumbs(), 1, 1, false)
		}
	} else {
		flex.RemoveItemAtIndex(2)
	}
}

func (a *App) buildHeader() tview.Primitive {
	header := tview.NewFlex()
	header.SetBackgroundColor(a.Styles.BgColor())
	header.SetDirection(tview.FlexColumn)
	if !a.showHeader {
		return header
	}

	clWidth := clusterInfoWidth
	if a.Conn().ConnectionOK() {
		n, err := a.Conn().Config().CurrentClusterName()
		if err == nil {
			size := len(n) + clusterInfoPad
			if size > clWidth {
				clWidth = size
			}
		}
	}
	header.AddItem(a.clusterInfo(), clWidth, 1, false)
	header.AddItem(a.Menu(), 0, 1, false)

	if a.showLogo {
		header.AddItem(a.Logo(), 26, 1, false)
	}

	return header
}

// Halt stop the application event loop.
func (a *App) Halt() {
	if a.cancelFn != nil {
		a.cancelFn()
		a.cancelFn = nil
	}
}

// Resume restarts the app event loop.
func (a *App) Resume() {
	var ctx context.Context
	ctx, a.cancelFn = context.WithCancel(context.Background())

	go a.clusterUpdater(ctx)

	if a.Config.K9s.UI.Reactive {
		if err := a.ConfigWatcher(ctx, a); err != nil {
			slog.Warn("ConfigWatcher failed", slogs.Error, err)
		}
		if err := a.SkinsDirWatcher(ctx, a); err != nil {
			slog.Warn("SkinsWatcher failed", slogs.Error, err)
		}
		if err := a.CustomViewsWatcher(ctx, a); err != nil {
			slog.Warn("CustomView watcher failed", slogs.Error, err)
		}
	}
}

func (a *App) clusterUpdater(ctx context.Context) {
	if err := a.refreshCluster(ctx); err != nil {
		slog.Error("Cluster updater failed!", slogs.Error, err)
		return
	}

	bf := model.NewExpBackOff(ctx, clusterRefresh, 2*time.Minute)
	delay := clusterRefresh
	for {
		select {
		case <-ctx.Done():
			slog.Debug("ClusterInfo updater canceled!")
			return
		case <-time.After(delay):
			if err := a.refreshCluster(ctx); err != nil {
				slog.Error("Cluster updates failed. Giving up ;(", slogs.Error, err)
				if delay = bf.NextBackOff(); delay == backoff.Stop {
					a.BailOut(1)
					return
				}
			} else {
				bf.Reset()
				delay = clusterRefresh
			}
		}
	}
}

func (a *App) refreshCluster(context.Context) error {
	c := a.Content.Top()
	if ok := a.Conn().CheckConnectivity(); ok {
		if atomic.LoadInt32(&a.conRetry) > 0 {
			atomic.StoreInt32(&a.conRetry, 0)
			a.Status(model.FlashInfo, "K8s connectivity OK")
			if c != nil {
				c.Start()
			}
		} else {
			a.ClearStatus(true)
		}
		a.factory.ValidatePortForwards()
	} else if c != nil {
		atomic.AddInt32(&a.conRetry, 1)
		c.Stop()
	}

	count, maxConnRetry := atomic.LoadInt32(&a.conRetry), a.Config.K9s.MaxConnRetry
	if count >= maxConnRetry {
		slog.Error("Conn check failed. Bailing out!",
			slogs.Retry, count,
			slogs.MaxRetries, maxConnRetry,
		)
		ExitStatus = fmt.Sprintf("Lost K8s connection (%d). Bailing out!", count)
		a.BailOut(1)
	}
	if count > 0 {
		a.Status(model.FlashWarn, fmt.Sprintf("Dial K8s Toast [%d/%d]", count, maxConnRetry))
		return fmt.Errorf("conn check failed (%d/%d)", count, maxConnRetry)
	}

	// Reload alias
	go func() {
		if err := a.command.Reset(a.Config.ContextAliasesPath(), false); err != nil {
			slog.Warn("Command reset failed", slogs.Error, err)
			a.QueueUpdateDraw(func() {
				a.Logo().Warn("Aliases load failed!")
			})
		}
	}()
	// Update cluster info
	a.clusterModel.Refresh()

	return nil
}

func (a *App) switchNS(ns string) error {
	if a.Config.ActiveNamespace() == ns {
		return nil
	}
	if ns == client.ClusterScope {
		ns = client.BlankNamespace
	}
	if err := a.Config.SetActiveNamespace(ns); err != nil {
		return err
	}

	return a.factory.SetActiveNS(ns)
}

func (a *App) switchContext(ci *cmd.Interpreter, force bool) error {
	contextName, ok := ci.HasContext()
	if (!ok || a.Config.ActiveContextName() == contextName) && !force {
		return nil
	}

	a.Halt()
	defer a.Resume()
	{
		a.Config.Reset()
		ct, err := a.Config.ActivateContext(contextName)
		if err != nil {
			return err
		}
		if cns, ok := ci.NSArg(); ok {
			ct.Namespace.Active = cns
		}

		p := cmd.NewInterpreter(a.Config.ActiveView())
		p.ResetContextArg()
		if p.IsContextCmd() {
			a.Config.SetActiveView(client.PodGVR.String())
		}
		ns := a.Config.ActiveNamespace()
		if !a.Conn().IsValidNamespace(ns) {
			slog.Warn("Unable to validate namespace", slogs.Namespace, ns)
			if err := a.Config.SetActiveNamespace(ns); err != nil {
				return err
			}
		}
		a.Flash().Infof("Using %q namespace", ns)

		if err := a.Config.Save(true); err != nil {
			slog.Error("Fail to save config to disk", slogs.Subsys, "config", slogs.Error, err)
		}
		a.initFactory(ns)
		if err := a.command.Reset(a.Config.ContextAliasesPath(), true); err != nil {
			return err
		}

		slog.Debug("Switching Context",
			slogs.Context, contextName,
			slogs.Namespace, ns,
			slogs.View, a.Config.ActiveView(),
		)
		a.Flash().Infof("Switching context to %q::%q", contextName, ns)
		a.ReloadStyles()
		a.gotoResource(a.Config.ActiveView(), "", true, true)
		a.clusterModel.Reset(a.factory)
	}

	return nil
}

func (a *App) initFactory(ns string) {
	a.factory.Terminate()
	a.factory.Start(ns)
}

// BailOut exists the application.
func (a *App) BailOut(exitCode int) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Bailout failed", slogs.Error, err)
		}
	}()

	if err := nukeK9sShell(a); err != nil {
		slog.Error("Unable to nuke k9s shell pod", slogs.Error, err)
	}

	a.stopImgScanner()
	a.factory.Terminate()
	a.App.BailOut(exitCode)
}

// Run starts the application loop.
func (a *App) Run() error {
	a.Resume()

	go func() {
		if !a.Config.K9s.IsSplashless() {
			<-time.After(splashDelay)
		}
		a.QueueUpdateDraw(func() {
			a.Main.SwitchToPage("main")
			// if command bar is already active, focus it
			if a.CmdBuff().IsActive() {
				a.SetFocus(a.Prompt())
			}
		})
	}()

	if err := a.command.defaultCmd(true); err != nil {
		return err
	}
	a.SetRunning(true)
	if err := a.Application.Run(); err != nil {
		return err
	}

	return nil
}

// Status reports a new app status for display.
func (a *App) Status(l model.FlashLevel, msg string) {
	a.QueueUpdateDraw(func() {
		if a.showHeader {
			a.setLogo(l, msg)
		} else {
			a.setIndicator(l, msg)
		}
	})
}

// IsBenchmarking check if benchmarks are active.
func (a *App) IsBenchmarking() bool {
	return a.Logo().IsBenchmarking()
}

// ClearStatus reset logo back to normal.
func (a *App) ClearStatus(flash bool) {
	a.QueueUpdate(func() {
		a.Logo().Reset()
		if flash {
			a.Flash().Clear()
		}
	})
}

func (a *App) setLogo(l model.FlashLevel, msg string) {
	switch l {
	case model.FlashErr:
		a.Logo().Err(msg)
	case model.FlashWarn:
		a.Logo().Warn(msg)
	case model.FlashInfo:
		a.Logo().Info(msg)
	default:
		a.Logo().Reset()
	}
}

func (a *App) setIndicator(l model.FlashLevel, msg string) {
	switch l {
	case model.FlashErr:
		a.statusIndicator().Err(msg)
	case model.FlashWarn:
		a.statusIndicator().Warn(msg)
	case model.FlashInfo:
		a.statusIndicator().Info(msg)
	default:
		a.statusIndicator().Reset()
	}
}

// PrevCmd pops the command stack.
func (a *App) PrevCmd(*tcell.EventKey) *tcell.EventKey {
	if !a.Content.IsLast() {
		a.Content.Pop()
	}

	return nil
}

func (a *App) toggleHeaderCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.Prompt().InCmdMode() {
		return evt
	}

	a.QueueUpdateDraw(func() {
		a.showHeader = !a.showHeader
		a.toggleHeader(a.showHeader, a.showLogo)
	})

	return nil
}

func (a *App) toggleCrumbsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.Prompt().InCmdMode() {
		return evt
	}

	a.QueueUpdateDraw(func() {
		a.showCrumbs = !a.showCrumbs
		a.toggleCrumbs(a.showCrumbs)
	})

	return nil
}

func (a *App) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.CmdBuff().IsActive() && !a.CmdBuff().Empty() {
		a.gotoResource(a.GetCmd(), "", true, true)
		a.ResetCmd()
		return nil
	}

	return evt
}

func (a *App) cowCmd(msg string) {
	d := a.Styles.Dialog()
	dialog.ShowError(&d, a.Content.Pages, msg)
}

func (a *App) dirCmd(path string, pushCmd bool) error {
	slog.Debug("Exec Dir command", slogs.Path, path)
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	if path == "." {
		dir, err := os.Getwd()
		if err == nil {
			path = dir
		}
	}
	if pushCmd {
		a.cmdHistory.Push("dir " + path)
	}

	return a.inject(NewDir(path), true)
}

func (a *App) quitCmd(evt *tcell.EventKey) *tcell.EventKey {
	noExit := a.Config.K9s.NoExitOnCtrlC
	if a.InCmdMode() {
		if isBailoutEvt(evt) && noExit {
			return nil
		}
		return evt
	}

	if !noExit {
		a.BailOut(0)
	}

	return nil
}

func (a *App) helpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if evt != nil && evt.Rune() == '?' && a.Prompt().InCmdMode() {
		return evt
	}

	top := a.Content.Top()
	if top != nil && top.Name() == "help" {
		a.Content.Pop()
		return nil
	}

	if err := a.inject(NewHelp(a), false); err != nil {
		a.Flash().Err(err)
	}

	a.Prompt().Deactivate()
	return nil
}

// previousCommand returns to the command prior to the current one in the history
func (a *App) previousCommand(evt *tcell.EventKey) *tcell.EventKey {
	if evt != nil && evt.Rune() == rune(ui.KeyLeftBracket) && a.Prompt().InCmdMode() {
		return evt
	}
	c, ok := a.cmdHistory.Back()
	if !ok {
		a.App.Flash().Warn("Can't go back any further")
		return evt
	}
	a.gotoResource(c, "", true, false)
	return nil
}

// nextCommand returns to the command subsequent to the current one in the history
func (a *App) nextCommand(evt *tcell.EventKey) *tcell.EventKey {
	if evt != nil && evt.Rune() == rune(ui.KeyRightBracket) && a.Prompt().InCmdMode() {
		return evt
	}
	c, ok := a.cmdHistory.Forward()
	if !ok {
		a.App.Flash().Warn("Can't go forward any further")
		return evt
	}
	// We go to the resource before updating the history so that
	// gotoResource doesn't add this command to the history
	a.gotoResource(c, "", true, false)
	return nil
}

// lastCommand switches between the last command and the current one a la `cd -`
func (a *App) lastCommand(evt *tcell.EventKey) *tcell.EventKey {
	if evt != nil && evt.Rune() == ui.KeyDash && a.Prompt().InCmdMode() {
		return evt
	}
	c, ok := a.cmdHistory.Top()
	if !ok {
		a.App.Flash().Warn("No previous view to switch to")
		return evt
	}
	a.gotoResource(c, "", true, false)

	return nil
}

func (a *App) aliasCmd(*tcell.EventKey) *tcell.EventKey {
	if a.Content.Top() != nil && a.Content.Top().Name() == aliasTitle {
		a.Content.Pop()
		return nil
	}

	if err := a.inject(NewAlias(client.AliGVR), false); err != nil {
		a.Flash().Err(err)
	}

	return nil
}

func (a *App) gotoResource(c, path string, clearStack, pushCmd bool) {
	err := a.command.run(cmd.NewInterpreter(c), path, clearStack, pushCmd)
	if err != nil {
		d := a.Styles.Dialog()
		dialog.ShowError(&d, a.Content.Pages, err.Error())
	}
}

func (a *App) inject(c model.Component, clearStack bool) error {
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	if err := c.Init(ctx); err != nil {
		slog.Error("Component init failed",
			slogs.Error, err,
			slogs.CompName, c.Name(),
		)
		return err
	}
	if clearStack {
		a.Content.Clear()
	}
	a.Content.Push(c)

	return nil
}

func (a *App) clusterInfo() *ClusterInfo {
	return a.Views()["clusterInfo"].(*ClusterInfo)
}

func (a *App) statusIndicator() *ui.StatusIndicator {
	return a.Views()["statusIndicator"].(*ui.StatusIndicator)
}
