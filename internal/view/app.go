package view

import (
	"context"
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
	devMode        = "dev"
	clusterRefresh = time.Duration(5 * time.Second)
	indicatorFmt   = "[orange::b]K9s [aqua::]%s [white::]%s:%s:%s [lawngreen::]%s%%[white::]::[darkturquoise::]%s%%"
)

// ActionsFunc augments Keybindinga.
type ActionsFunc func(ui.KeyActions)

type focusHandler func(tview.Primitive)

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
	a.Content.Pop()

	return nil
}

func (a *App) Init(version string, rate int) {
	ctx := context.WithValue(context.Background(), ui.KeyApp, a)
	a.Content.Init(ctx)
	a.Content.Stack.AddListener(a.Crumbs())

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

	// ctx := context.WithValue(context.Background(), ui.KeyApp, a)
	// a.Content.Init(ctx)
	// d := NewDetails(a, nil)
	// d.SetText("Fuck!!")
	// a.Content.Push(d)
	// d = NewDetails(a, nil)
	// d.SetText("Shit!!")
	// a.Content.Push(d)

	main.AddItem(a.indicator(), 1, 1, false)
	main.AddItem(a.Content, 0, 10, true)
	main.AddItem(a.Crumbs(), 2, 1, false)
	main.AddItem(a.Flash(), 2, 1, false)
	a.toggleHeader(!a.Config.K9s.GetHeadless())
}

// func (a *App) StackPushed(c model.Component) {
// 	ctx := context.WithValue(context.Background(), ui.KeyApp, a)
// 	ctx, a.cancelFn = context.WithCancel(context.Background())
// 	c.Init(ctx)

// 	a.Frame().AddPage(c.Name(), c, true, true)
// 	a.SetFocus(c)
// 	a.setHints(c.Hints())
// }

// func (a *App) StackPopped(o, c model.Component) {
// 	a.Frame().RemovePage(o.Name())
// 	if c != nil {
// 		a.StackPushed(c)
// 	}
// }

// func (a *App) StackTop(model.Component) {
// }

// Changed indicates the buffer was changed.
func (a *App) BufferChanged(s string) {}

// Active indicates the buff activity changed.
func (a *App) BufferActive(state bool, _ ui.BufferKind) {
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
	flex := a.Main.GetPrimitive("main").(*tview.Flex)
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
		log.Debug().Msgf("Namespace did not change %s", ns)
		return true
	}
	a.Config.SetActiveNamespace(ns)

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
	a.startInformer(ns)
	a.Config.Reset()
	a.Config.Save()
	a.Flash().Infof("Switching context to %s", ctx)
	if load {
		a.gotoResource("po", true)
	}

	return nil
}

func (a *App) startInformer(ns string) bool {
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
		a.gotoResource(a.GetCmd(), true)
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

func (a *App) gotoResource(res string, record bool) bool {
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
