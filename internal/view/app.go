package view

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	splashTime     = 1
	clusterRefresh = time.Duration(5 * time.Second)
	indicatorFmt   = "[orange::b]K9s [aqua::]%s [white::]%s:%s:%s [lawngreen::]%s%%[white::]::[darkturquoise::]%s%%"
)

// App represents an application view.
type App struct {
	*ui.App

	Content    *PageStack
	command    *command
	factory    *watch.Factory
	forwarders model.Forwarders
	version    string
	showHeader bool
	cancelFn   context.CancelFunc
}

// NewApp returns a K9s app instance.
func NewApp(cfg *config.Config) *App {
	v := App{
		App:        ui.NewApp(),
		Content:    NewPageStack(),
		forwarders: model.NewForwarders(),
	}
	v.Config = cfg
	v.InitBench(cfg.K9s.CurrentCluster)
	v.command = newCommand(&v)

	v.Views()["indicator"] = ui.NewIndicatorView(v.App, v.Styles)
	v.Views()["clusterInfo"] = newClusterInfoView(&v, k8s.NewMetricsServer(cfg.GetConnection()))

	return &v
}

// ActiveView returns the currently active view.
func (a *App) ActiveView() model.Component {
	return a.Content.GetPrimitive("main").(model.Component)
}

func (a *App) PrevCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !a.Content.IsLast() {
		a.Content.Pop()
	}

	return nil
}

func (a *App) Init(version string, rate int) error {
	ctx := context.WithValue(context.Background(), ui.KeyApp, a)
	if err := a.Content.Init(ctx); err != nil {
		return err
	}
	a.Content.Stack.AddListener(a.Crumbs())
	a.Content.Stack.AddListener(a.Menu())

	a.version = version
	a.App.Init()
	a.CmdBuff().AddListener(a)
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

	a.clusterInfo().init(version)
	if a.Config.K9s.GetHeadless() {
		a.refreshIndicator()
	}

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(a.indicator(), 1, 1, false)
	main.AddItem(a.Content, 0, 10, true)
	main.AddItem(a.Crumbs(), 2, 1, false)
	main.AddItem(a.Flash(), 2, 1, false)

	a.Main.AddPage("main", main, true, false)
	a.Main.AddPage("splash", ui.NewSplash(a.Styles, version), true, true)
	a.toggleHeader(!a.Config.K9s.GetHeadless())

	if err := a.command.Init(); err != nil {
		panic(err)
	}

	return nil
}

func (a *App) bindKeys() {
	a.AddActions(ui.KeyActions{
		ui.KeyH:        ui.NewKeyAction("ToggleHeader", a.toggleHeaderCmd, false),
		ui.KeyHelp:     ui.NewKeyAction("Help", a.helpCmd, false),
		tcell.KeyCtrlA: ui.NewKeyAction("Aliases", a.aliasCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Goto", a.gotoCmd, false),
	})
}

// Changed indicates the buffer was changed.
func (a *App) BufferChanged(s string) {}

// Active indicates the buff activity changed.
func (a *App) BufferActive(state bool, _ ui.BufferKind) {
	flex, ok := a.Main.GetPrimitive("main").(*tview.Flex)
	if !ok {
		return
	}

	if state && flex.ItemAt(1) != a.Cmd() {
		flex.AddItemAtIndex(1, a.Cmd(), 3, 1, false)
	} else if !state && flex.ItemAt(1) == a.Cmd() {
		flex.RemoveItemAtIndex(1)
	}
	a.Draw()
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
		flex.AddItemAtIndex(0, a.indicator(), 1, 1, false)
		a.refreshIndicator()
	}
}

func (a *App) buildHeader() tview.Primitive {
	header := tview.NewFlex()
	header.SetBorderPadding(0, 0, 1, 1)
	header.SetDirection(tview.FlexColumn)
	if !a.showHeader {
		return header
	}
	header.AddItem(a.clusterInfo(), 35, 1, false)
	header.AddItem(a.Menu(), 0, 1, false)
	header.AddItem(a.Logo(), 26, 1, false)

	return header
}

func (a *App) Halt() {
	if a.cancelFn != nil {
		a.cancelFn()
	}
}

func (a *App) Resume() {
	var ctx context.Context
	ctx, a.cancelFn = context.WithCancel(context.Background())
	go a.clusterUpdater(ctx)
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

func (a *App) refreshClusterInfo() {
	log.Debug().Msgf("***** REFRESHING CLUSTER ******")
	if !a.showHeader {
		a.refreshIndicator()
	} else {
		a.clusterInfo().refresh()
	}
}

func (a *App) refreshIndicator() {
	mx := k8s.NewMetricsServer(a.Conn())
	cluster := resource.NewCluster(a.Conn(), &log.Logger, mx)
	var cmx k8s.ClusterMetrics
	nos, nmx, err := fetchResources(a)
	cpu, mem := "0", "0"
	if err == nil {
		cluster.Metrics(nos, nmx, &cmx)
		cpu = resource.AsPerc(cmx.PercCPU)
		if cpu == "0" {
			cpu = resource.NAValue
		}
		mem = resource.AsPerc(cmx.PercMEM)
		if mem == "0" {
			mem = resource.NAValue
		}
	}

	info := fmt.Sprintf(
		indicatorFmt,
		a.version,
		cluster.ClusterName(),
		cluster.UserName(),
		cluster.Version(),
		cpu,
		mem,
	)
	a.indicator().SetPermanent(info)
}

func (a *App) switchNS(ns string) bool {
	if ns == resource.AllNamespace {
		ns = resource.AllNamespaces
	}
	if err := a.Config.SetActiveNamespace(ns); err != nil {
		log.Error().Err(err).Msg("Config Set NS failed!")
		return false
	}
	a.factory.SetActive(ns)

	return true
}

func (a *App) switchCtx(name string, loadPods bool) error {
	log.Debug().Msgf("Switching Context %q", name)

	a.Halt()
	defer a.Resume()
	{
		a.forwarders.DeleteAll()
		ns, err := a.Conn().Config().CurrentNamespaceName()
		if err != nil {
			log.Info().Err(err).Msg("No namespace specified in context. Using K9s config")
		}
		a.initFactory(ns)

		a.Config.Reset()
		if err := a.Config.Save(); err != nil {
			log.Error().Err(err).Msg("Config save failed!")
		}
		a.Flash().Infof("Switching context to %s", name)
		if loadPods && !a.gotoResource("pods") {
			a.Flash().Err(errors.New("Goto pods failed"))
		}
		a.refreshClusterInfo()
	}

	return nil
}

func (a *App) initFactory(ns string) {
	a.factory.Terminate()
	a.factory.Init()
	a.factory.SetActive(ns)
}

// BailOut exists the application.
func (a *App) BailOut() {
	a.factory.Terminate()
	a.forwarders.DeleteAll()
	a.App.BailOut()
}

// Run starts the application loop
func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a.Halt()

	// Only enable skin updater while in dev mode.
	if a.HasSkins {
		if err := a.StylesUpdater(ctx, a); err != nil {
			log.Error().Err(err).Msg("Unable to track skin changes")
		}
	}

	go func() {
		<-time.After(splashTime * time.Second)
		a.QueueUpdateDraw(func() {
			a.Main.SwitchToPage("main")
		})
	}()

	a.command.defaultCmd()
	if err := a.Application.Run(); err != nil {
		panic(err)
	}
}

func (a *App) status(l ui.FlashLevel, msg string) {
	a.Flash().Info(msg)
	a.setIndicator(l, msg)
	a.setLogo(l, msg)
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
		a.indicator().Err(msg)
	case ui.FlashWarn:
		a.indicator().Warn(msg)
	case ui.FlashInfo:
		a.indicator().Info(msg)
	default:
		a.indicator().Reset()
	}
	a.Draw()
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
		if !a.gotoResource(a.GetCmd()) {
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
	} else {
		a.inject(NewHelp())
	}

	return nil
}

func (a *App) aliasCmd(evt *tcell.EventKey) *tcell.EventKey {
	if _, ok := a.Content.GetPrimitive("main").(*Alias); ok {
		return evt
	}

	if a.Content.Top() != nil && a.Content.Top().Name() == aliasTitle {
		a.Content.Pop()
	} else {
		a.inject(NewAlias())
	}

	return nil
}

func (a *App) gotoResource(res string) bool {
	return a.command.run(res)
}

func (a *App) inject(c model.Component) {
	a.Content.Push(c)
}

func (a *App) clusterInfo() *clusterInfoView {
	return a.Views()["clusterInfo"].(*clusterInfoView)
}

func (a *App) indicator() *ui.IndicatorView {
	return a.Views()["indicator"].(*ui.IndicatorView)
}
