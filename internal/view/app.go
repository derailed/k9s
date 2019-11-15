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
	"k8s.io/client-go/tools/portforward"
)

const (
	splashTime     = 1
	clusterRefresh = time.Duration(5 * time.Second)
	indicatorFmt   = "[orange::b]K9s [aqua::]%s [white::]%s:%s:%s [lawngreen::]%s%%[white::]::[darkturquoise::]%s%%"
)

// ActionsFunc augments Keybindinga.
type ActionsFunc func(ui.KeyActions)

type forwarder interface {
	Start(path, co string, ports []string) (*portforward.PortForwarder, error)
	Stop()
	Path() string
	Container() string
	Ports() []string
	Active() bool
	Age() string
}

// ResourceViewer represents a generic resource viewer.
type ResourceViewer interface {
	model.Component

	setEnterFn(enterFn)
	setColorerFn(ui.ColorerFunc)
	setDecorateFn(decorateFn)
	setExtraActionsFn(ActionsFunc)
	masterPage() *Table
}

// App represents an application view.
type App struct {
	*ui.App

	Content    *PageStack
	command    *command
	informer   *watch.Informer
	stopCh     chan struct{}
	forwarders map[string]forwarder
	version    string
	showHeader bool
}

// NewApp returns a K9s app instance.
func NewApp(cfg *config.Config) *App {
	v := App{
		App:        ui.NewApp(),
		Content:    NewPageStack(),
		forwarders: make(map[string]forwarder),
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
	log.Debug().Msgf("------ CONTENT PREVIOUS")
	a.Content.DumpStack()
	a.Content.DumpPages()
	if !a.Content.IsLast() {
		a.Content.Pop()
	}

	return nil
}

func (a *App) Init(version string, rate int) {
	ctx := context.WithValue(context.Background(), ui.KeyApp, a)
	a.Content.Init(ctx)
	a.Content.Stack.AddListener(a.Crumbs())
	a.Content.Stack.AddListener(a.Menu())

	a.version = version
	a.CmdBuff().AddListener(a)
	a.App.Init()
	a.AddActions(ui.KeyActions{
		ui.KeyH:        ui.NewKeyAction("ToggleHeader", a.toggleHeaderCmd, false),
		ui.KeyHelp:     ui.NewKeyAction("Help", a.helpCmd, false),
		tcell.KeyCtrlA: ui.NewKeyAction("Aliases", a.aliasCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("Goto", a.gotoCmd, false),
	})

	if a.Conn() != nil {
		ns, err := a.Conn().Config().CurrentNamespaceName()
		if err != nil {
			log.Info().Msg("No namespace specified using all namespaces")
		}
		a.startInformer(ns)
		a.clusterInfo().init(version)
		if a.Config.K9s.GetHeadless() {
			a.refreshIndicator()
		}
	}

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	a.Main.AddPage("main", main, true, false)
	a.Main.AddPage("splash", ui.NewSplash(a.Styles, version), true, true)

	main.AddItem(a.indicator(), 1, 1, false)
	main.AddItem(a.Content, 0, 10, true)
	main.AddItem(a.Crumbs(), 2, 1, false)
	main.AddItem(a.Flash(), 2, 1, false)
	a.toggleHeader(!a.Config.K9s.GetHeadless())
}

// Changed indicates the buffer was changed.
func (a *App) BufferChanged(s string) {}

// Active indicates the buff activity changed.
func (a *App) BufferActive(state bool, _ ui.BufferKind) {
	log.Debug().Msgf("App Buffer Activated!")
	flex, ok := a.Main.GetPrimitive("main").(*tview.Flex)
	if !ok {
		return
	}
	if state {
		flex.AddItemAtIndex(1, a.Cmd(), 3, 1, false)
	} else if flex.ItemAt(1) == a.Cmd() {
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

func (a *App) clusterUpdater(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("Cluster updater canceled!")
			return
		case <-time.After(clusterRefresh):
			a.QueueUpdateDraw(func() {
				if !a.showHeader {
					a.refreshIndicator()
				} else {
					a.clusterInfo().refresh()
				}
			})
		}
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
	if ns == a.Config.ActiveNamespace() {
		return true
	}

	if err := a.Config.SetActiveNamespace(ns); err != nil {
		log.Error().Err(err).Msg("Config Set NS failed!")
	}

	return a.startInformer(ns)
}

func (a *App) switchCtx(ctx string, load bool) error {
	l := resource.NewContext(a.Conn())
	if err := l.Switch(ctx); err != nil {
		return err
	}

	a.stopForwarders()
	ns, err := a.Conn().Config().CurrentNamespaceName()
	if err != nil {
		log.Info().Err(err).Msg("No namespace specified using all namespaces")
	}
	if a.stopCh != nil {
		close(a.stopCh)
		a.stopCh = nil
	}
	a.informer = nil
	a.startInformer(ns)
	a.Config.Reset()
	if err := a.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config save failed!")
	}
	a.Flash().Infof("Switching context to %s", ctx)
	if load && !a.gotoResource("po") {
		a.Flash().Err(errors.New("Goto pod failed"))
	}

	return nil
}

func (a *App) startInformer(ns string) bool {
	// if informer watches all ns - don't start a new informer then.
	if a.informer != nil && a.informer.Namespace == resource.AllNamespaces {
		log.Debug().Msgf(">>>> Informer is already watching all namespaces. No restart needed ;)")
		return true
	}

	if a.stopCh != nil {
		close(a.stopCh)
		a.stopCh = nil
	}

	var err error
	a.informer, err = watch.NewInformer(a.Conn(), ns)
	if err != nil {
		log.Error().Err(err).Msgf("%v", err)
		a.Flash().Err(err)
		return false
	}
	a.stopCh = make(chan struct{})
	a.informer.Run(a.stopCh)
	if a.Config.K9s.GetHeadless() {
		a.refreshIndicator()
	}

	return true
}

// BailOut exists the application.
func (a *App) BailOut() {
	if a.stopCh != nil {
		log.Debug().Msg("<<<< Stopping Watcher")
		close(a.stopCh)
		a.stopCh = nil
	}

	a.stopForwarders()
	a.App.BailOut()
}

func (a *App) stopForwarders() {
	for k, f := range a.forwarders {
		log.Debug().Msgf("Deleting forwarder %s", f.Path())
		f.Stop()
		delete(a.forwarders, k)
	}
}

// Run starts the application loop
func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.clusterUpdater(ctx)

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
	if a.Config.K9s.GetHeadless() {
		a.setIndicator(l, msg)
	} else {
		a.setLogo(l, msg)
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
		a.Content.Stack.Reset()
		if !a.gotoResource(a.GetCmd()) {
			a.Flash().Errf("Goto %s failed!", a.GetCmd())
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
	a.inject(NewHelp())

	return nil
}

func (a *App) aliasCmd(evt *tcell.EventKey) *tcell.EventKey {
	if _, ok := a.Content.GetPrimitive("main").(*Alias); ok {
		return evt
	}
	a.inject(NewAlias())

	return nil
}

func (a *App) gotoResource(res string) bool {
	return a.command.run(res)
}

func (a *App) inject(c model.Component) {
	log.Debug().Msgf("Injecting component %#v", c)
	a.Content.Push(c)
}

func (a *App) clusterInfo() *clusterInfoView {
	return a.Views()["clusterInfo"].(*clusterInfoView)
}

func (a *App) indicator() *ui.IndicatorView {
	return a.Views()["indicator"].(*ui.IndicatorView)
}
