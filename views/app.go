package views

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
	log "github.com/sirupsen/logrus"
)

const (
	splashTime = 1
	defaultNS  = ""
)

type (
	focusHandler func(tview.Primitive)

	igniter interface {
		tview.Primitive
		init(ctx context.Context, ns string)
	}

	resourceViewer interface {
		igniter
	}

	appView struct {
		*tview.Application

		refreshRate  int
		version      string
		defaultNS    string
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
		cmdBuff      []rune
		isCmd        bool
	}
)

// NewApp returns a K9s app instance.
func NewApp(v string, rate int, ns string) *appView {
	var app appView
	{
		app = appView{
			Application: tview.NewApplication(),
			pages:       tview.NewPages(),
			version:     v,
			menuView:    newMenuView(),
			content:     tview.NewPages(),
			refreshRate: rate,
			defaultNS:   ns,
		}
		app.command = newCommand(&app)
		app.focusChanged = app.changedFocus
		app.SetInputCapture(app.keyboard)
	}
	return &app
}

func (a *appView) Init() {
	a.infoView = newInfoView(a)
	a.infoView.init()

	a.flashView = newFlashView(a.Application, "Initializing...")

	header := tview.NewFlex()
	{
		header.SetDirection(tview.FlexColumn)
		header.AddItem(a.infoView, 25, 1, false)
		header.AddItem(a.menuView, 0, 1, false)
		header.AddItem(logoView(), 26, 1, false)
	}

	main := tview.NewFlex()
	{
		main.SetDirection(tview.FlexRow)
		main.AddItem(header, 7, 1, false)
		main.AddItem(a.content, 0, 10, true)
		main.AddItem(a.flashView, 1, 1, false)
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
		log.Panic(err)
	}
}

func (a *appView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	switch evt.Rune() {
	case 'q':
		a.quit(evt)
		return nil
	}

	switch evt.Key() {
	case tcell.KeyCtrlR:
		a.Draw()
	case tcell.KeyEsc:
		a.resetCmd()
	case tcell.KeyEnter:
		if len(a.cmdBuff) != 0 {
			a.command.run(string(a.cmdBuff))
		}
		a.resetCmd()
	case tcell.KeyTab:
		a.nextFocus()
	case tcell.KeyRune:
		if evt.Rune() == ':' {
			a.isCmd = true
		} else if a.isCmd {
			a.cmdBuff = append([]rune(a.cmdBuff), evt.Rune())
		}
	}
	return evt
}

func (a *appView) resetCmd() {
	a.cmdBuff = []rune{}
	a.isCmd = false
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

	{
		var ctx context.Context
		ctx, a.cancel = context.WithCancel(context.TODO())
		p.init(ctx, a.defaultNS)
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

func (a *appView) refresh() {
	a.infoView.refresh()
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
