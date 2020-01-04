package view

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	splashTime         = 1
	clusterRefresh     = time.Duration(5 * time.Second)
	statusIndicatorFmt = "[orange::b]K9s [aqua::]%s [white::]%s:%s:%s [lawngreen::]%s%%[white::]::[darkturquoise::]%s%%"
)

// App represents an application view.
type App struct {
	*ui.App

	Content    *PageStack
	command    *Command
	factory    *watch.Factory
	version    string
	showHeader bool
	cancelFn   context.CancelFunc
}

// NewApp returns a K9s app instance.
func NewApp(cfg *config.Config) *App {
	a := App{
		App:     ui.NewApp(cfg.K9s.CurrentCluster),
		Content: NewPageStack(),
	}
	a.Config = cfg
	a.InitBench(cfg.K9s.CurrentCluster)

	a.Views()["statusIndicator"] = ui.NewStatusIndicator(a.App, a.Styles)
	a.Views()["clusterInfo"] = NewClusterInfo(&a, client.NewMetricsServer(cfg.GetConnection()))

	return &a
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

	a.command = NewCommand(a)
	if err := a.command.Init(); err != nil {
		return err
	}

	a.clusterInfo().Init(version)
	if a.Config.K9s.GetHeadless() {
		a.refreshIndicator()
	}

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
		ui.KeyH:        ui.NewSharedKeyAction("ToggleHeader", a.toggleHeaderCmd, false),
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
		a.refreshIndicator()
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
	header.AddItem(a.clusterInfo(), 40, 1, false)
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
			log.Debug().Msg("Cluster updater canceled!")
			return
		case <-time.After(clusterRefresh):
			a.QueueUpdateDraw(func() {
				a.refreshClusterInfo()
			})
		}
	}
}

// BOZO!! Refact to use model/view strategy.
func (a *App) refreshClusterInfo() {
	if !a.showHeader {
		a.refreshIndicator()
	} else {
		a.clusterInfo().refresh()
	}
}

func (a *App) refreshIndicator() {
	mx := client.NewMetricsServer(a.Conn())
	cluster := model.NewCluster(a.Conn(), mx)
	var cmx client.ClusterMetrics
	nos, nmx, err := fetchResources(a)
	if err != nil {
		log.Error().Err(err).Msgf("unable to refresh cluster statusIndicator")
		return
	}

	if err := cluster.Metrics(nos, nmx, &cmx); err != nil {
		log.Error().Err(err).Msgf("unable to refresh cluster statusIndicator")
		return
	}

	cpu := render.AsPerc(cmx.PercCPU)
	if cpu == "0" {
		cpu = render.NAValue
	}
	mem := render.AsPerc(cmx.PercMEM)
	if mem == "0" {
		mem = render.NAValue
	}

	a.statusIndicator().SetPermanent(fmt.Sprintf(
		statusIndicatorFmt,
		a.version,
		cluster.ClusterName(),
		cluster.UserName(),
		cluster.Version(),
		cpu,
		mem,
	))
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

		if err := a.command.Reset(); err != nil {
			return err
		}
		a.Config.Reset()
		if err := a.Config.Save(); err != nil {
			log.Error().Err(err).Msg("Config save failed!")
		}
		a.Flash().Infof("Switching context to %s", name)
		if err := a.gotoResource("pods", true); loadPods && err != nil {
			a.Flash().Err(err)
		}
		a.refreshClusterInfo()
		a.ReloadStyles(name)
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
func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a.Halt()

	if err := a.StylesUpdater(ctx, a); err != nil {
		log.Error().Err(err).Msg("Unable to track skin changes")
	}

	go func() {
		<-time.After(splashTime * time.Second)
		a.QueueUpdateDraw(func() {
			a.Main.SwitchToPage("main")
		})
	}()

	if err := a.command.defaultCmd(); err != nil {
		panic(err)
	}
	if err := a.Application.Run(); err != nil {
		panic(err)
	}
}

// Status reports a new app status for display.
func (a *App) Status(l ui.FlashLevel, msg string) {
	a.Flash().SetMessage(l, msg)
	a.setIndicator(l, msg)
	a.setLogo(l, msg)
	a.Draw()
}

// ClearStatus reset log back to normal.
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
			return nil
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

func (a *App) gotoResource(res string, clearStack bool) error {
	return a.command.run(res, clearStack)
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
