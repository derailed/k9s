package views

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
	log "github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const splashTime = 1

type (
	focusHandler func(tview.Primitive)

	igniter interface {
		tview.Primitive
		getTitle() string
		init(ctx context.Context, ns string)
	}

	resourceViewer interface {
		igniter
	}

	appView struct {
		*tview.Application

		version      string
		pages        *tview.Pages
		content      *tview.Pages
		flashView    *flashView
		menuView     *menuView
		infoView     *infoView
		command      *command
		focusGroup   []tview.Primitive
		focusCurrent int
		focusChanged focusHandler
		cancel       context.CancelFunc
		cmdBuff      *cmdBuff
		cmdView      *cmdView
	}
)

// NewApp returns a K9s app instance.
func NewApp() *appView {
	var app appView
	{
		app = appView{
			Application: tview.NewApplication(),
			pages:       tview.NewPages(),
			menuView:    newMenuView(),
			content:     tview.NewPages(),
			cmdBuff:     newCmdBuff(':'),
			cmdView:     newCmdView('🐶'),
		}
		app.command = newCommand(&app)
		app.focusChanged = app.changedFocus
		app.SetInputCapture(app.keyboard)
	}
	return &app
}

func (a *appView) Init(v string, rate int, flags *genericclioptions.ConfigFlags) {
	a.version = v
	a.infoView = newInfoView(a)
	a.infoView.init()

	a.flashView = newFlashView(a.Application, "Initializing...")

	a.cmdBuff.addListener(a.cmdView)

	header := tview.NewFlex()
	{
		header.SetDirection(tview.FlexColumn)
		header.AddItem(a.infoView, 55, 1, false)
		header.AddItem(a.menuView, 0, 1, false)
		header.AddItem(logoView(), 26, 1, false)
	}

	main := tview.NewFlex()
	{
		main.SetDirection(tview.FlexRow)
		main.AddItem(header, 7, 1, false)
		main.AddItem(a.cmdView, 1, 1, false)
		main.AddItem(a.content, 0, 10, true)
		main.AddItem(a.flashView, 2, 1, false)
	}

	a.pages.AddPage("main", main, true, false)
	a.pages.AddPage("splash", NewSplash(a.version), true, true)
	a.SetRoot(a.pages, true)
}

// Run starts the application loop
func (a *appView) Run() {
	go func() {
		<-time.After(splashTime * time.Second)
		a.showPage("main")
		a.Draw()
	}()

	a.command.defaultCmd()
	if err := a.Application.Run(); err != nil {
		panic(err)
	}
}

func (a *appView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		switch evt.Rune() {
		case 'q':
			if !a.cmdMode() {
				a.quit(evt)
			}
		case '?':
			if !a.cmdView.inCmdMode() {
				a.help(evt)
				return evt
			}
		case a.cmdBuff.hotKey:
			a.flash(flashInfo, "Entering command mode...")
			log.Debug("K9s entering command mode...")
			a.cmdBuff.setActive(true)
			a.cmdBuff.clear()
			return evt
		}

		if a.cmdBuff.isActive() {
			a.cmdBuff.add(evt.Rune())
		}
		return evt
	}

	switch evt.Key() {
	case tcell.KeyCtrlR:
		log.Debug("Refreshing screen...")
		a.Draw()
	case tcell.KeyEsc:
		a.cmdBuff.reset()
	case tcell.KeyEnter:
		if a.cmdBuff.isActive() && !a.cmdBuff.empty() {
			a.command.run(a.cmdBuff.String())
		}
		a.cmdBuff.setActive(false)
	case tcell.KeyBackspace2:
		if a.cmdBuff.isActive() {
			a.cmdBuff.del()
		}
	case tcell.KeyTab:
		a.nextFocus()
	}
	return evt
}

func (a *appView) showPage(p string) {
	a.pages.SwitchToPage(p)
}

func (a *appView) inject(p igniter) {
	if a.cancel != nil {
		a.cancel()
	}
	a.content.RemovePage("main")
	a.content.AddPage("main", p, true, true)

	var ctx context.Context
	{
		ctx, a.cancel = context.WithCancel(context.TODO())
		p.init(ctx, config.Root.ActiveNamespace())
	}

	go func() {
		<-time.After(100 * time.Millisecond)
		a.Draw()
	}()

	a.focusGroup = append([]tview.Primitive{}, p)
	a.focusCurrent = 0
	a.fireFocusChanged(p)
	a.SetFocus(p)
}

func (a *appView) cmdMode() bool {
	return a.cmdView.inCmdMode()
}

func (a *appView) refresh() {
	a.infoView.refresh()
}

func (a *appView) help(*tcell.EventKey) {
	a.inject(newHelpView(a))
}

func (a *appView) noop(*tcell.EventKey) {
}

func (a *appView) quit(*tcell.EventKey) {
	a.Stop()
	os.Exit(0)
}

func (a *appView) flash(level flashLevel, m ...string) {
	a.flashView.setMessage(level, m...)
}

func (a *appView) setHints(h hints) {
	a.menuView.setMenu(h)
}

func logoView() tview.Primitive {
	v := tview.NewTextView()
	{
		v.SetWordWrap(false)
		v.SetWrap(false)
		v.SetDynamicColors(true)
		for i, s := range logoSmall {
			fmt.Fprintf(v, "[orange::b]%s", s)
			if i+1 < len(logoSmall) {
				fmt.Fprintf(v, "\n")
			}
		}
	}
	return v
}

func (a *appView) fireFocusChanged(p tview.Primitive) {
	if a.focusChanged != nil {
		a.focusChanged(p)
	}
}

func (a *appView) changedFocus(p tview.Primitive) {
	switch p.(type) {
	case hinter:
		a.setHints(p.(hinter).hints())
	}
}

func (a *appView) nextFocus() {
	for i := range a.focusGroup {
		if i == a.focusCurrent {
			a.focusCurrent = 0
			victim := a.focusGroup[a.focusCurrent]
			if i+1 < len(a.focusGroup) {
				a.focusCurrent++
				victim = a.focusGroup[a.focusCurrent]
			}
			a.fireFocusChanged(victim)
			a.SetFocus(victim)
			return
		}
	}
	return
}
