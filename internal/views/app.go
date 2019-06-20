package views

import (
	"context"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
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

	igniter interface {
		tview.Primitive

		init(ctx context.Context, ns string)
	}

	resourceViewer interface {
		igniter

		setEnterFn(enterFn)
		setColorerFn(colorerFn)
		setDecorateFn(decorateFn)
		setExtraActionsFn(actionsFn)
	}

	appView struct {
		*shellView

		cmdBuff    *cmdBuff
		command    *command
		cancel     context.CancelFunc
		informer   *watch.Informer
		stopCh     chan struct{}
		forwarders []forwarder
	}
)

// NewApp returns a K9s app instance.
func NewApp(cfg *config.Config) *appView {
	v := appView{
		shellView: newShellView(),
		cmdBuff:   newCmdBuff(':'),
	}

	v.config = cfg
	v.initBench(cfg.K9s.CurrentCluster)
	v.refreshStyles()

	v.views["menu"] = newMenuView(v.styles)
	v.views["logo"] = newLogoView(v.styles)
	v.views["cmd"] = newCmdView(v.styles, '🐶')
	v.command = newCommand(&v)
	v.views["flash"] = newFlashView(&v, "Initializing...")
	v.views["crumbs"] = newCrumbsView(v.styles)
	v.views["clusterInfo"] = newClusterInfoView(&v, k8s.NewMetricsServer(cfg.GetConnection()))

	v.actions = keyActions{
		KeyColon:            newKeyAction("Cmd", v.activateCmd, false),
		tcell.KeyCtrlR:      newKeyAction("Redraw", v.redrawCmd, false),
		tcell.KeyCtrlC:      newKeyAction("Quit", v.quitCmd, false),
		KeyHelp:             newKeyAction("Help", v.helpCmd, false),
		tcell.KeyCtrlA:      newKeyAction("Aliases", v.aliasCmd, true),
		tcell.KeyEscape:     newKeyAction("Escape", v.escapeCmd, false),
		tcell.KeyEnter:      newKeyAction("Goto", v.gotoCmd, false),
		tcell.KeyBackspace2: newKeyAction("Erase", v.eraseCmd, false),
		tcell.KeyBackspace:  newKeyAction("Erase", v.eraseCmd, false),
		tcell.KeyDelete:     newKeyAction("Erase", v.eraseCmd, false),
	}
	v.SetInputCapture(v.keyboard)

	return &v
}

func (a *appView) Init(version string, rate int) {
	if a.conn() != nil {
		ns, err := a.conn().Config().CurrentNamespaceName()
		if err != nil {
			log.Info().Err(err).Msg("No namespace specified using all namespaces")
		}
		a.startInformer(ns)
		a.clusterInfo().init(version)
	}
	a.cmdBuff.addListener(a.cmd())

	header := tview.NewFlex()
	{
		header.SetDirection(tview.FlexColumn)
		header.AddItem(a.clusterInfo(), 35, 1, false)
		header.AddItem(a.views["menu"], 0, 1, false)
		header.AddItem(a.logo(), 26, 1, false)
	}

	main := tview.NewFlex()
	{
		main.SetDirection(tview.FlexRow)
		main.AddItem(header, 7, 1, false)
		main.AddItem(a.cmd(), 3, 1, false)
		main.AddItem(a.content, 0, 10, true)
		main.AddItem(a.crumbs(), 2, 1, false)
		main.AddItem(a.flash(), 1, 1, false)
	}

	a.pages.AddPage("main", main, true, false)
	a.pages.AddPage("splash", newSplash(a.styles, version), true, true)
	a.SetRoot(a.pages, true)
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
	a.informer, err = watch.NewInformer(a.conn(), ns)
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
	a.Stop()
}

func (a *appView) stopForwarders() {
	for _, f := range a.forwarders {
		log.Debug().Msgf("Deleting forwarder %s", f.Path())
		f.Stop()
	}
	a.forwarders = []forwarder{}
}

func (a *appView) conn() k8s.Connection {
	return a.config.GetConnection()
}

// Run starts the application loop
func (a *appView) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.clusterUpdater(ctx)

	// Only enable skin updater while in dev mode.
	if a.hasSkins {
		if err := a.stylesUpdater(ctx, a); err != nil {
			log.Error().Err(err).Msg("Unable to track skin changes")
		}
	}

	go func() {
		<-time.After(splashTime * time.Second)
		a.QueueUpdateDraw(func() {
			a.pages.SwitchToPage("main")
		})
	}()

	a.command.defaultCmd()
	if err := a.Application.Run(); err != nil {
		panic(err)
	}
}

func (a *appView) statusReset() {
	a.logo().reset()
	a.Draw()
}

func (a *appView) status(l flashLevel, msg string) {
	a.flash().info(msg)

	switch l {
	case flashErr:
		a.logo().err(msg)
	case flashWarn:
		a.logo().warn(msg)
	case flashInfo:
		a.logo().info(msg)
	default:
		a.logo().reset()
	}
	a.Draw()
}

func (a *appView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if a.cmdBuff.isActive() && evt.Modifiers() == tcell.ModNone {
			a.cmdBuff.add(evt.Rune())
			return nil
		}
		key = tcell.Key(evt.Rune())
		if evt.Modifiers() == tcell.ModAlt {
			key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
		}
	}

	if a, ok := a.actions[key]; ok {
		log.Debug().Msgf(">> AppView handled key: %s", tcell.KeyNames[key])
		return a.action(evt)
	}

	return evt
}

func (a *appView) redrawCmd(evt *tcell.EventKey) *tcell.EventKey {
	a.Draw()
	return evt
}

func (a *appView) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.cmdBuff.isActive() {
		a.cmdBuff.del()
		return nil
	}
	return evt
}

func (a *appView) escapeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.cmdBuff.isActive() {
		a.cmdBuff.reset()
	}
	return evt
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
	if a.cmdBuff.isActive() && !a.cmdBuff.empty() {
		a.gotoResource(a.cmdBuff.String(), true)
		a.cmdBuff.reset()
		return nil
	}
	a.cmdBuff.setActive(false)
	return evt
}

func (a *appView) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.inCmdMode() {
		return evt
	}
	a.flash().info("Command mode activated.")
	a.cmdBuff.setActive(true)
	a.cmdBuff.clear()
	return nil
}

func (a *appView) quitCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.inCmdMode() {
		return evt
	}
	a.BailOut()

	return nil
}

func (a *appView) helpCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.inCmdMode() {
		return evt
	}
	a.inject(newHelpView(a, a.currentView()))
	return nil
}

func (a *appView) currentView() igniter {
	return a.content.GetPrimitive("main").(igniter)
}

func (a *appView) aliasCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.inCmdMode() {
		return evt
	}
	a.inject(newAliasView(a, a.currentView()))

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

func (a *appView) inject(i igniter) {
	if a.cancel != nil {
		a.cancel()
	}
	a.content.RemovePage("main")
	var ctx context.Context
	{
		ctx, a.cancel = context.WithCancel(context.Background())
		i.init(ctx, a.config.ActiveNamespace())
	}
	a.content.AddPage("main", i, true, true)
	a.SetFocus(i)
}

func (a *appView) inCmdMode() bool {
	return a.cmd().inCmdMode()
}

func (a *appView) setHints(h hints) {
	a.views["menu"].(*menuView).populateMenu(h)
}

// View Accessors...

func (a *appView) crumbs() *crumbsView {
	return a.views["crumbs"].(*crumbsView)
}

func (a *appView) logo() *logoView {
	return a.views["logo"].(*logoView)
}

func (a *appView) clusterInfo() *clusterInfoView {
	return a.views["clusterInfo"].(*clusterInfoView)
}

func (a *appView) flash() *flashView {
	return a.views["flash"].(*flashView)
}

func (a *appView) cmd() *cmdView {
	return a.views["cmd"].(*cmdView)
}
