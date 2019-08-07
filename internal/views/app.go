package views

import (
	"context"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
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
	clusterRefresh = time.Duration(15 * time.Second)
)

type (
	focusHandler func(tview.Primitive)

	forwarder interface {
		Start(path, co string, ports []string) (*portforward.PortForwarder, error)
		Stop()
		Path() string
		Container() string
		Ports() []string
		Active() bool
		Age() string
	}

	resourceViewer interface {
		ui.Igniter

		setEnterFn(enterFn)
		setColorerFn(ui.ColorerFunc)
		setDecorateFn(decorateFn)
		setExtraActionsFn(ui.ActionsFunc)
	}

	appView struct {
		*ui.App

		command    *command
		cancel     context.CancelFunc
		informer   *watch.Informer
		stopCh     chan struct{}
		forwarders map[string]forwarder
	}
)

// NewApp returns a K9s app instance.
func NewApp(cfg *config.Config) *appView {
	v := appView{
		App:        ui.NewApp(),
		forwarders: make(map[string]forwarder),
	}
	v.Config = cfg
	v.InitBench(cfg.K9s.CurrentCluster)
	v.command = newCommand(&v)

	v.Views()["flash"] = ui.NewFlashView(v.Application, "Initializing...")
	v.Views()["clusterInfo"] = newClusterInfoView(&v, k8s.NewMetricsServer(cfg.GetConnection()))

	return &v
}

func (a *appView) Init(version string, rate int) {
	a.App.Init()

	a.AddActions(ui.KeyActions{
		ui.KeyHelp:     ui.NewKeyAction("Help", a.helpCmd, false),
		tcell.KeyCtrlA: ui.NewKeyAction("Aliases", a.aliasCmd, true),
		tcell.KeyEnter: ui.NewKeyAction("Goto", a.gotoCmd, false),
	})

	if a.Conn() != nil {
		ns, err := a.Conn().Config().CurrentNamespaceName()
		if err != nil {
			log.Info().Msg("No namespace specified using all namespaces")
		}
		a.startInformer(ns)
		a.clusterInfo().init(version)
	}

	header := tview.NewFlex()
	{
		header.SetDirection(tview.FlexColumn)
		header.AddItem(a.clusterInfo(), 35, 1, false)
		header.AddItem(a.Menu(), 0, 1, false)
		header.AddItem(a.Logo(), 26, 1, false)
	}

	main := tview.NewFlex()
	{
		main.SetDirection(tview.FlexRow)
		main.AddItem(header, 7, 1, false)
		main.AddItem(a.Cmd(), 3, 1, false)
		main.AddItem(a.Frame(), 0, 10, true)
		main.AddItem(a.Crumbs(), 2, 1, false)
		main.AddItem(a.Flash(), 1, 1, false)
	}

	a.Main().AddPage("main", main, true, false)
	a.Main().AddPage("splash", ui.NewSplash(a.Styles, version), true, true)
}

func (a *appView) clusterUpdater(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("Cluster updater canceled!")
			return
		case <-time.After(clusterRefresh):
			a.QueueUpdateDraw(func() {
				a.clusterInfo().refresh()
			})
		}
	}
}

func (a *appView) startInformer(ns string) {
	if a.stopCh != nil {
		close(a.stopCh)
	}

	var err error
	a.stopCh = make(chan struct{})
	a.informer, err = watch.NewInformer(a.Conn(), ns)
	if err != nil {
		log.Panic().Err(err).Msgf("%v", err)
	}
	a.informer.Run(a.stopCh)
}

// BailOut exists the application.
func (a *appView) BailOut() {
	if a.stopCh != nil {
		log.Debug().Msg("<<<< Stopping Watcher")
		close(a.stopCh)
		a.stopCh = nil
	}

	if a.cancel != nil {
		a.cancel()
	}
	a.stopForwarders()
	a.App.BailOut()
}

func (a *appView) stopForwarders() {
	for k, f := range a.forwarders {
		log.Debug().Msgf("Deleting forwarder %s", f.Path())
		f.Stop()
		delete(a.forwarders, k)
	}
}

// Run starts the application loop
func (a *appView) Run() {
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
			a.Main().SwitchToPage("main")
		})
	}()

	a.command.defaultCmd()
	if err := a.Application.Run(); err != nil {
		panic(err)
	}
}

func (a *appView) status(l ui.FlashLevel, msg string) {
	a.Flash().Info(msg)
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

func (a *appView) prevCmd(evt *tcell.EventKey) *tcell.EventKey {
	if top, ok := a.command.previousCmd(); ok {
		log.Debug().Msgf("Previous command %s", top)
		a.gotoResource(top, false)
		return nil
	}
	return evt
}

func (a *appView) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.GetCmdBuff().IsActive() && !a.GetCmdBuff().Empty() {
		a.gotoResource(a.GetCmd(), true)
		a.ResetCmd()
		return nil
	}
	a.ActivateCmd(false)

	return evt
}

func (a *appView) helpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.InCmdMode() {
		return evt
	}
	a.inject(newHelpView(a, a.ActiveView()))
	return nil
}

func (a *appView) aliasCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.InCmdMode() {
		return evt
	}
	a.inject(newAliasView(a, a.ActiveView()))

	return nil
}

func (a *appView) gotoResource(res string, record bool) bool {
	if a.cancel != nil {
		a.cancel()
	}
	valid := a.command.run(res)
	if valid && record {
		a.command.pushCmd(res)
	}

	return valid
}

func (a *appView) inject(i ui.Igniter) {
	if a.cancel != nil {
		a.cancel()
	}
	a.Frame().RemovePage("main")
	var ctx context.Context
	{
		ctx, a.cancel = context.WithCancel(context.Background())
		i.Init(ctx, a.Config.ActiveNamespace())
	}
	a.Frame().AddPage("main", i, true, true)
	a.SetFocus(i)
}

func (a *appView) clusterInfo() *clusterInfoView {
	return a.Views()["clusterInfo"].(*clusterInfoView)
}
