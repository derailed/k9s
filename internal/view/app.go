package view

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
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
)

// ExitStatus indicates UI exit conditions.
var ExitStatus = ""

const (
	splashDelay      = 1 * time.Second
	clusterRefresh   = 5 * time.Second
	maxConRetry      = 10
	clusterInfoWidth = 50
	clusterInfoPad   = 15
)

// App represents an application view.
type App struct {
	*ui.App

	Content      *PageStack
	command      *Command
	factory      *watch.Factory
	version      string
	showHeader   bool
	cancelFn     context.CancelFunc
	conRetry     int32
	clusterModel *model.ClusterInfo
}

// NewApp returns a K9s app instance.
func NewApp(cfg *config.Config) *App {
	a := App{
		App:     ui.NewApp(cfg.K9s.CurrentContext),
		Content: NewPageStack(),
	}
	a.Config = cfg
	a.InitBench(cfg.K9s.CurrentCluster)

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
	a.bindKeys()
	if a.Conn() == nil {
		return errors.New("No client connection detected")
	}
	ns, err := a.Conn().Config().CurrentNamespaceName()
	if err != nil {
		log.Info().Msg("No namespace specified using all namespaces")
	}

	a.factory = watch.NewFactory(a.Conn())
	a.initFactory(ns)

	a.clusterModel = model.NewClusterInfo(a.factory, version)
	a.clusterModel.AddListener(a.clusterInfo())
	a.clusterModel.AddListener(a.statusIndicator())
	a.clusterModel.Refresh()

	a.command = NewCommand(a)
	if err := a.command.Init(); err != nil {
		return err
	}

	a.clusterInfo().Init()

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(a.statusIndicator(), 1, 1, false)
	main.AddItem(a.Content, 0, 10, true)
	main.AddItem(a.Crumbs(), 2, 1, false)
	main.AddItem(a.Flash(), 2, 1, false)

	a.Main.AddPage("main", main, true, false)
	a.Main.AddPage("splash", ui.NewSplash(a.Styles, version), true, true)
	a.toggleHeader(!a.Config.K9s.GetHeadless())

	return nil
}

func (a *App) bindKeys() {
	a.AddActions(ui.KeyActions{
		tcell.KeyCtrlH: ui.NewSharedKeyAction("ToggleHeader", a.toggleHeaderCmd, false),
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

func (a *App) buildHeader() tview.Primitive {
	header := tview.NewFlex()
	header.SetBackgroundColor(a.Styles.BgColor())
	header.SetBorderPadding(0, 0, 1, 1)
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
	}
}

// Resume restarts the app event loop.
func (a *App) Resume() {
	var ctx context.Context
	ctx, a.cancelFn = context.WithCancel(context.Background())
	go a.clusterUpdater(ctx)
	if err := a.StylesUpdater(ctx, a); err != nil {
		log.Error().Err(err).Msgf("Styles update failed")
	}
}

func (a *App) clusterUpdater(ctx context.Context) {
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
			a.Status(ui.FlashInfo, "K8s connectivity OK")
			if c != nil {
				c.Start()
			}
		}
	} else {
		atomic.AddInt32(&a.conRetry, 1)
		if c != nil {
			c.Stop()
		}
		count := atomic.LoadInt32(&a.conRetry)
		log.Warn().Msgf("Conn check failed (%d/%d)", count, maxConRetry)
		a.Status(ui.FlashWarn, fmt.Sprintf("Dial K8s failed (%d)", count))
	}

	count := atomic.LoadInt32(&a.conRetry)
	if count >= maxConRetry {
		ExitStatus = fmt.Sprintf("Lost K8s connection (%d). Bailing out!", count)
		a.BailOut()
	}
	if count > 0 {
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

func (a *App) switchNS(ns string) bool {
	if ns == client.ClusterScope {
		ns = client.AllNamespaces
	}
	if err := a.Config.SetActiveNamespace(ns); err != nil {
		log.Error().Err(err).Msg("Config Set NS failed!")
		return false
	}
	a.factory.SetActiveNS(ns)

	return true
}

func (a *App) switchCtx(name string, loadPods bool) error {
	log.Debug().Msgf("Switching Context %q", name)

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
		a.Config.Reset()
		if err := a.Config.Save(); err != nil {
			log.Error().Err(err).Msg("Config save failed!")
		}
		a.Flash().Infof("Switching context to %s", name)
		a.ReloadStyles(name)
		if err := a.gotoResource("pods", true); loadPods && err != nil {
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
	if err := a.Application.Run(); err != nil {
		return err
	}

	return nil
}

// Status reports a new app status for display.
func (a *App) Status(l ui.FlashLevel, msg string) {
	a.Flash().SetMessage(l, msg)
	a.setIndicator(l, msg)
	a.setLogo(l, msg)
	a.Draw()
}

// ClearStatus reset logo back to normal.
func (a *App) ClearStatus(flash bool) {
	a.Logo().Reset()
	if flash {
		a.Flash().Clear()
	}
	a.Draw()
}

func (a *App) setLogo(l ui.FlashLevel, msg string) {
	switch l {
	case ui.FlashErr:
		a.Logo().Err(msg)
	case ui.FlashWarn:
		a.Logo().Warn(msg)
	case ui.FlashInfo:
		a.Logo().Info(msg)
	default:
		a.Logo().Reset()
	}
	a.Draw()
}

func (a *App) setIndicator(l ui.FlashLevel, msg string) {
	switch l {
	case ui.FlashErr:
		a.statusIndicator().Err(msg)
	case ui.FlashWarn:
		a.statusIndicator().Warn(msg)
	case ui.FlashInfo:
		a.statusIndicator().Info(msg)
	default:
		a.statusIndicator().Reset()
	}
	a.Draw()
}

// PrevCmd pops the command stack.
func (a *App) PrevCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !a.Content.IsLast() {
		a.Content.Pop()
	}

	return nil
}

func (a *App) toggleHeaderCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.Cmd().InCmdMode() {
		return evt
	}

	a.showHeader = !a.showHeader
	a.toggleHeader(a.showHeader)
	a.Draw()

	return nil
}

func (a *App) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.CmdBuff().IsActive() && !a.CmdBuff().Empty() {
		if err := a.gotoResource(a.GetCmd(), true); err != nil {
			log.Error().Err(err).Msgf("Goto resource for %q failed", a.GetCmd())
			a.Flash().Err(err)
		}
		a.ResetCmd()
		return nil
	}
	a.ActivateCmd(false)

	return evt
}

func (a *App) helpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if _, ok := a.Content.GetPrimitive("main").(*Help); ok {
		return evt
	}
	if a.Content.Top() != nil && a.Content.Top().Name() == helpTitle {
		a.Content.Pop()
		return nil
	}

	if err := a.inject(NewHelp()); err != nil {
		a.Flash().Err(err)
	}

	return nil
}

func (a *App) aliasCmd(evt *tcell.EventKey) *tcell.EventKey {
	if _, ok := a.Content.GetPrimitive("main").(*Alias); ok {
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

func (a *App) viewResource(gvr, path string, clearStack bool) error {
	return a.command.run(gvr, path, clearStack)
}

func (a *App) gotoResource(cmd string, clearStack bool) error {
	return a.command.run(cmd, "", clearStack)
}

func (a *App) inject(c model.Component) error {
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	if err := c.Init(ctx); err != nil {
		return fmt.Errorf("component init failed for %q %v", c.Name(), err)
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
