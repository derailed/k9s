package view

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExitStatus indicates UI exit conditions.
var ExitStatus = ""

const (
	splashDelay      = 1 * time.Second
	clusterRefresh   = 15 * time.Second
	maxConRetry      = 15
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
	showCrumbs    bool
}

// NewApp returns a K9s app instance.
func NewApp(cfg *config.Config) *App {
	a := App{
		App:           ui.NewApp(cfg, cfg.K9s.CurrentContext),
		cmdHistory:    model.NewHistory(model.MaxHistory),
		filterHistory: model.NewHistory(model.MaxHistory),
		Content:       NewPageStack(),
	}

	a.Views()["statusIndicator"] = ui.NewStatusIndicator(a.App, a.Styles)
	a.Views()["clusterInfo"] = NewClusterInfo(&a)

	return &a
}

// ConOK checks the connection is cool, returns false otherwise.
func (a *App) ConOK() bool {
	return atomic.LoadInt32(&a.conRetry) == 0
}

// Init initializes the application.
func (a *App) Init(version string, rate int) error {
	a.version = version

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
		return errors.New("No client connection detected")
	}
	ns, err := a.Conn().Config().CurrentNamespaceName()
	log.Debug().Msgf("CURRENT-NS %q -- %v", ns, err)
	if err != nil {
		log.Info().Msg("No namespace specified using cluster default namespace")
	} else if err = a.Config.SetActiveNamespace(ns); err != nil {
		log.Error().Err(err).Msgf("Fail to set active namespace to %q", ns)
	}

	a.factory = watch.NewFactory(a.Conn())
	ok, err := a.isValidNS(ns)
	if !ok && err == nil {
		return fmt.Errorf("Invalid namespace %s", ns)
	}
	a.initFactory(ns)

	a.clusterModel = model.NewClusterInfo(a.factory, version)
	a.clusterModel.AddListener(a.clusterInfo())
	a.clusterModel.AddListener(a.statusIndicator())
	a.clusterModel.Refresh()
	a.clusterInfo().Init()

	a.command = NewCommand(a)
	if err := a.command.Init(); err != nil {
		return err
	}
	a.CmdBuff().SetSuggestionFn(a.suggestCommand())
	// BOZO!!
	// a.CmdBuff().AddListener(a)

	a.layout(ctx, version)
	a.initSignals()

	return nil
}

func (a *App) layout(ctx context.Context, version string) {
	flash := ui.NewFlash(a.App)
	go flash.Watch(ctx, a.Flash().Channel())

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(a.statusIndicator(), 1, 1, false)
	main.AddItem(a.Content, 0, 10, true)
	main.AddItem(flash, 1, 1, false)

	a.Main.AddPage("main", main, true, false)
	a.Main.AddPage("splash", ui.NewSplash(a.Styles, version), true, true)
	a.toggleHeader(!a.Config.K9s.GetHeadless())
	a.toggleCrumbs(!a.Config.K9s.GetCrumbsless())
}

func (a *App) initSignals() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)

	go func(sig chan os.Signal) {
		s := <-sig
		if s == syscall.SIGHUP {
			os.Exit(0)
		}
	}(sig)
}

func (a *App) suggestCommand() model.SuggestionFunc {
	return func(s string) (entries sort.StringSlice) {
		if s == "" {
			if a.cmdHistory.Empty() {
				return
			}
			return a.cmdHistory.List()
		}

		s = strings.ToLower(s)
		for _, k := range a.command.alias.Aliases.Keys() {
			if k == s {
				continue
			}
			if strings.HasPrefix(k, s) {
				entries = append(entries, strings.Replace(k, s, "", 1))
			}
		}
		if len(entries) == 0 {
			return nil
		}
		entries.Sort()
		return
	}
}

func (a *App) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if k, ok := a.HasAction(ui.AsKey(evt)); ok && !a.Content.IsTopDialog() {
		return k.Action(evt)
	}

	return evt
}

func (a *App) bindKeys() {
	a.AddActions(ui.KeyActions{
		tcell.KeyCtrlE: ui.NewSharedKeyAction("ToggleHeader", a.toggleHeaderCmd, false),
		tcell.KeyCtrlT: ui.NewSharedKeyAction("toggleCrumbs", a.toggleCrumbsCmd, false),
		ui.KeyHelp:     ui.NewSharedKeyAction("Help", a.helpCmd, false),
		tcell.KeyCtrlA: ui.NewSharedKeyAction("Aliases", a.aliasCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Goto", a.gotoCmd, false),
	})
}

// ActiveView returns the currently active view.
func (a *App) ActiveView() model.Component {
	return a.Content.GetPrimitive("main").(model.Component)
}

func (a *App) toggleHeader(flag bool) {
	a.showHeader = flag
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
		flex.AddItemAtIndex(2, a.Crumbs(), 1, 1, false)
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
	n, err := a.Conn().Config().CurrentClusterName()
	if err == nil {
		size := len(n) + clusterInfoPad
		if size > clWidth {
			clWidth = size
		}
	}
	header.AddItem(a.clusterInfo(), clWidth, 1, false)
	header.AddItem(a.Menu(), 0, 1, false)
	header.AddItem(a.Logo(), 26, 1, false)

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
	if err := a.StylesWatcher(ctx, a); err != nil {
		log.Error().Err(err).Msgf("Styles watcher failed")
	}
	if err := a.CustomViewsWatcher(ctx, a); err != nil {
		log.Error().Err(err).Msgf("CustomView watcher failed")
	}
}

func (a *App) clusterUpdater(ctx context.Context) {
	a.refreshCluster()
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("ClusterInfo updater canceled!")
			return
		case <-time.After(clusterRefresh):
			a.refreshCluster()
		}
	}
}

func (a *App) refreshCluster() {
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

	count := atomic.LoadInt32(&a.conRetry)
	if count >= maxConRetry {
		ExitStatus = fmt.Sprintf("Lost K8s connection (%d). Bailing out!", count)
		a.BailOut()
	}
	if count > 0 {
		log.Warn().Msgf("Conn check failed (%d/%d)", count, maxConRetry)
		a.Status(model.FlashWarn, fmt.Sprintf("Dial K8s failed (%d)", count))
		return
	}

	// Reload alias
	go func() {
		if err := a.command.Reset(false); err != nil {
			log.Error().Err(err).Msgf("Command reset failed")
		}
	}()

	// Update cluster info
	a.clusterModel.Refresh()
}

func (a *App) switchNS(ns string) error {
	if ns == client.ClusterScope {
		ns = client.AllNamespaces
	}
	ok, err := a.isValidNS(ns)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Invalid namespace %q", ns)
	}
	if err := a.Config.SetActiveNamespace(ns); err != nil {
		return fmt.Errorf("Unable to save active namespace in config")
	}

	return a.factory.SetActiveNS(ns)
}

func (a *App) isValidNS(ns string) (bool, error) {
	if ns == client.AllNamespaces || ns == client.NamespaceAll {
		return true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.Conn().Config().CallTimeout())
	defer cancel()
	dial, err := a.Conn().Dial()
	if err != nil {
		return false, err
	}
	_, err = dial.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
	if err != nil {
		log.Warn().Err(err).Msgf("Validation failed for namespace: %q", ns)
	}

	return true, nil
}

func (a *App) switchCtx(name string, loadPods bool) error {
	log.Debug().Msgf("--> Switching Context %q--%q", name, a.Config.ActiveView())
	a.Halt()
	defer a.Resume()
	{
		ns, err := a.Conn().Config().CurrentNamespaceName()
		if err != nil {
			log.Warn().Msg("No namespace specified in context. Using K9s config")
		}
		a.initFactory(ns)

		if err := a.command.Reset(true); err != nil {
			return err
		}
		v := a.Config.ActiveView()
		if v == "" || isContextCmd(v) || loadPods {
			v = "pod"
			a.Config.SetActiveView(v)
		}
		if err := a.Config.Save(); err != nil {
			log.Error().Err(err).Msg("Config save failed!")
		}
		a.Config.Reset()

		a.Flash().Infof("Switching context to %s", name)
		a.ReloadStyles(name)
		if err := a.gotoResource(v, "", true); loadPods && err != nil {
			a.Flash().Err(err)
		}
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

	if err := nukeK9sShell(a); err != nil {
		log.Error().Err(err).Msgf("nuking k9s shell pod")
	}
	a.factory.Terminate()
	a.App.BailOut()
}

// Run starts the application loop
func (a *App) Run() error {
	a.Resume()

	go func() {
		<-time.After(splashDelay)
		a.QueueUpdateDraw(func() {
			a.Main.SwitchToPage("main")
		})
	}()

	if err := a.command.defaultCmd(); err != nil {
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
		a.toggleHeader(a.showHeader)
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
		if err := a.gotoResource(a.GetCmd(), "", true); err != nil {
			log.Error().Err(err).Msgf("Goto resource for %q failed", a.GetCmd())
			a.Flash().Err(err)
		}
		a.ResetCmd()
		return nil
	}

	return evt
}

func (a *App) meowCmd(msg string) {
	if err := a.inject(NewMeow(a, msg)); err != nil {
		a.Flash().Err(err)
	}
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
	a.Content.Stack.Clear()
	a.cmdHistory.Push("dir " + path)

	return a.inject(NewDir(path))
}

func (a *App) helpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.CmdBuff().InCmdMode() {
		return evt
	}

	if a.Content.Top() != nil && a.Content.Top().Name() == "help" {
		a.Content.Pop()
		return nil
	}

	if err := a.inject(NewHelp()); err != nil {
		a.Flash().Err(err)
	}

	return nil
}

func (a *App) aliasCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.CmdBuff().InCmdMode() {
		return evt
	}

	if a.Content.Top() != nil && a.Content.Top().Name() == aliasTitle {
		a.Content.Pop()
		return nil
	}

	if err := a.inject(NewAlias(client.NewGVR("aliases"))); err != nil {
		a.Flash().Err(err)
	}

	return nil
}

func (a *App) gotoResource(cmd, path string, clearStack bool) error {
	err := a.command.run(cmd, path, clearStack)
	if err == nil {
		return err
	}

	c := NewMeow(a, err.Error())
	_ = c.Init(context.Background())
	if clearStack {
		a.Content.Stack.Clear()
	}
	a.Content.Push(c)

	return nil
}

func (a *App) inject(c model.Component) error {
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	if err := c.Init(ctx); err != nil {
		log.Error().Err(err).Msgf("component init failed for %q %v", c.Name(), err)
		c = NewMeow(a, err.Error())
		_ = c.Init(ctx)
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
