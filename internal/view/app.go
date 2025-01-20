// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/k9s/internal/vul"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
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
func (a *App) Init(version string, rate int) error {
	a.version = model.NormalizeVersion(version)

	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	if err := a.Content.Init(ctx); err != nil {
		return err
	}
	a.Content.Stack.AddListener(a.Crumbs())
	a.Content.Stack.AddListener(a.Menu())

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

func (a *App) stopImgScanner() {
	if vul.ImgScanner != nil {
		vul.ImgScanner.Stop()
	}
}

func (a *App) initImgScanner(version string) {
	defer func(t time.Time) {
		log.Debug().Msgf("Scanner init time %s", time.Since(t))
	}(time.Now())

	vul.ImgScanner = vul.NewImageScanner(a.Config.K9s.ImageScans)
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
	a.Main.AddPage("splash", ui.NewSplash(a.Styles, a.version), true, true)
	a.toggleHeader(!a.Config.K9s.IsHeadless(), !a.Config.K9s.IsLogoless())
}

func (a *App) initSignals() {
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
		log.Error().Err(err).Msg("failed to list contexts")
	}

	return func(s string) (entries sort.StringSlice) {
		if s == "" {
			if a.cmdHistory.Empty() {
				return
			}
			return a.cmdHistory.List()
		}

		ls := strings.ToLower(s)
		for _, k := range a.command.alias.Aliases.Keys() {
			if suggest, ok := cmd.ShouldAddSuggest(ls, k); ok {
				entries = append(entries, suggest)
			}
		}

		namespaceNames, err := a.factory.Client().ValidNamespaceNames()
		if err != nil {
			log.Error().Err(err).Msg("failed to list namespaces")
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
	if !a.Conn().ConnectionOK() {
		return nil, errors.New("no connection")
	}
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
		ui.KeyShift9:   ui.NewSharedKeyAction("DumpGOR", a.dumpGOR, false),
		tcell.KeyCtrlE: ui.NewSharedKeyAction("ToggleHeader", a.toggleHeaderCmd, false),
		tcell.KeyCtrlG: ui.NewSharedKeyAction("toggleCrumbs", a.toggleCrumbsCmd, false),
		ui.KeyHelp:     ui.NewSharedKeyAction("Help", a.helpCmd, false),
		tcell.KeyCtrlA: ui.NewSharedKeyAction("Aliases", a.aliasCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Goto", a.gotoCmd, false),
		tcell.KeyCtrlC: ui.NewKeyAction("Quit", a.quitCmd, false),
	}))
}

func (a *App) dumpGOR(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msgf("GOR %d", runtime.NumGoroutine())
	// bb := make([]byte, 5_000_000)
	// runtime.Stack(bb, true)
	// log.Debug().Msgf("GOR\n%s", string(bb))
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
		log.Fatal().Msg("Expecting valid flex view")
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
		log.Fatal().Msg("Expecting valid flex view")
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
			log.Warn().Err(err).Msgf("ConfigWatcher failed")
		}
		if err := a.SkinsDirWatcher(ctx, a); err != nil {
			log.Warn().Err(err).Msgf("SkinsWatcher failed")
		}
		if err := a.CustomViewsWatcher(ctx, a); err != nil {
			log.Warn().Err(err).Msgf("CustomView watcher failed")
		}
	}
}

func (a *App) clusterUpdater(ctx context.Context) {
	if err := a.refreshCluster(ctx); err != nil {
		log.Error().Err(err).Msgf("Cluster updater failed!")
		return
	}

	bf := model.NewExpBackOff(ctx, clusterRefresh, 2*time.Minute)
	delay := clusterRefresh
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("ClusterInfo updater canceled!")
			return
		case <-time.After(delay):
			if err := a.refreshCluster(ctx); err != nil {
				log.Error().Err(err).Msgf("ClusterUpdater failed")
				if delay = bf.NextBackOff(); delay == backoff.Stop {
					a.BailOut()
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

	count, maxConnRetry := atomic.LoadInt32(&a.conRetry), int32(a.Config.K9s.MaxConnRetry)
	if count >= maxConnRetry {
		log.Error().Msgf("Conn check failed (%d/%d). Bailing out!", count, maxConnRetry)
		ExitStatus = fmt.Sprintf("Lost K8s connection (%d). Bailing out!", count)
		a.BailOut()
	}
	if count > 0 {
		a.Status(model.FlashWarn, fmt.Sprintf("Dial K8s Toast [%d/%d]", count, maxConnRetry))
		return fmt.Errorf("conn check failed (%d/%d)", count, maxConnRetry)
	}

	// Reload alias
	go func() {
		if err := a.command.Reset(a.Config.ContextAliasesPath(), false); err != nil {
			log.Warn().Err(err).Msgf("Command reset failed")
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
	name, ok := ci.HasContext()
	if !ok || a.Config.ActiveContextName() == name {
		if !force {
			return nil
		}
	}

	a.Halt()
	defer a.Resume()
	{
		a.Config.Reset()
		ct, err := a.Config.K9s.ActivateContext(name)
		if err != nil {
			return err
		}
		if cns, ok := ci.NSArg(); ok {
			ct.Namespace.Active = cns
		}

		p := cmd.NewInterpreter(a.Config.ActiveView())
		p.ResetContextArg()
		if p.IsContextCmd() {
			a.Config.SetActiveView("pod")
		}
		ns := a.Config.ActiveNamespace()
		if !a.Conn().IsValidNamespace(ns) {
			log.Warn().Msgf("Unable to validate namespace: %q. Using %q as active namespace", ns, ns)
			if err := a.Config.SetActiveNamespace(ns); err != nil {
				return err
			}
		}
		a.Flash().Infof("Using %q namespace", ns)

		if err := a.Config.Save(true); err != nil {
			log.Error().Err(err).Msg("config save failed!")
		} else {
			log.Debug().Msgf("Saved context config for: %q", name)
		}
		a.initFactory(ns)
		if err := a.command.Reset(a.Config.ContextAliasesPath(), true); err != nil {
			return err
		}

		log.Debug().Msgf("--> Switching Context %q -- %q -- %q", name, ns, a.Config.ActiveView())
		a.Flash().Infof("Switching context to %q::%q", name, ns)
		a.ReloadStyles()
		a.gotoResource(a.Config.ActiveView(), "", true)
		a.clusterModel.Reset(a.factory)
	}

	return nil
}

func (a *App) initFactory(ns string) {
	a.factory.Terminate()
	a.factory.Start(ns)
}

// BailOut exists the application.
func (a *App) BailOut() {
	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Bailing out %v", err)
		}
	}()

	if err := a.Config.Save(true); err != nil {
		log.Error().Err(err).Msg("config save failed!")
	}

	if err := nukeK9sShell(a); err != nil {
		log.Error().Err(err).Msgf("nuking k9s shell pod")
	}

	a.stopImgScanner()
	a.factory.Terminate()
	a.App.BailOut()
}

// Run starts the application loop.
func (a *App) Run() error {
	a.Resume()

	go func() {
		<-time.After(splashDelay)
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
func (a *App) PrevCmd(evt *tcell.EventKey) *tcell.EventKey {
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
		a.gotoResource(a.GetCmd(), "", true)
		a.ResetCmd()
		return nil
	}

	return evt
}

func (a *App) cowCmd(msg string) {
	dialog.ShowError(a.Styles.Dialog(), a.Content.Pages, msg)
}

func (a *App) dirCmd(path string) error {
	log.Debug().Msgf("DIR PATH %q", path)
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
	a.cmdHistory.Push("dir " + path)

	return a.inject(NewDir(path), true)
}

func (a *App) quitCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.InCmdMode() {
		return evt
	}

	if !a.Config.K9s.NoExitOnCtrlC {
		a.BailOut()
	}

	// overwrite the default ctrl-c behavior of tview
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

func (a *App) aliasCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.Content.Top() != nil && a.Content.Top().Name() == aliasTitle {
		a.Content.Pop()
		return nil
	}

	if err := a.inject(NewAlias(client.NewGVR("aliases")), false); err != nil {
		a.Flash().Err(err)
	}

	return nil
}

func (a *App) gotoResource(c, path string, clearStack bool) {
	err := a.command.run(cmd.NewInterpreter(c), path, clearStack)
	if err != nil {
		dialog.ShowError(a.Styles.Dialog(), a.Content.Pages, err.Error())
	}
}

func (a *App) inject(c model.Component, clearStack bool) error {
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	if err := c.Init(ctx); err != nil {
		log.Error().Err(err).Msgf("Component init failed for %q", c.Name())
		return err
	}
	if clearStack {
		a.Content.Stack.Clear()
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
