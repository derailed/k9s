package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type (
	keyHandler interface {
		keyboard(evt *tcell.EventKey) *tcell.EventKey
	}

	actionsFn func(keyActions)

	shellView struct {
		*tview.Application
		configurator

		actions keyActions
		pages   *tview.Pages
		content *tview.Pages
		views   map[string]tview.Primitive
		cmdBuff *cmdBuff
	}
)

func newShellView() *shellView {
	s := shellView{
		Application: tview.NewApplication(),
		actions:     make(keyActions),
		pages:       tview.NewPages(),
		content:     tview.NewPages(),
		views:       make(map[string]tview.Primitive),
		cmdBuff:     newCmdBuff(':'),
	}

	s.refreshStyles()

	s.views["menu"] = newMenuView(s.styles)
	s.views["logo"] = newLogoView(s.styles)
	s.views["cmd"] = newCmdView(s.styles, '🐶')
	s.views["crumbs"] = newCrumbsView(s.styles)

	return &s
}

func (s *shellView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if s.cmdBuff.isActive() && evt.Modifiers() == tcell.ModNone {
			s.cmdBuff.add(evt.Rune())
			return nil
		}
		key = asKey(evt)
	}

	if a, ok := s.actions[key]; ok {
		log.Debug().Msgf(">> AppView handled key: %s", tcell.KeyNames[key])
		return a.action(evt)
	}

	return evt
}

func (s *shellView) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.cmdBuff.isActive() {
		s.cmdBuff.del()
		return nil
	}
	return evt
}

func (s *shellView) escapeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.cmdBuff.isActive() {
		s.cmdBuff.reset()
	}
	return evt
}

func (s *shellView) conn() k8s.Connection {
	return s.config.GetConnection()
}

func (s *shellView) init() {
}

func (s *shellView) redrawCmd(evt *tcell.EventKey) *tcell.EventKey {
	s.Draw()
	return evt
}

func (s *shellView) currentView() igniter {
	return s.content.GetPrimitive("main").(igniter)
}

func (s *shellView) setHints(h hints) {
	s.views["menu"].(*menuView).populateMenu(h)
}

func (s *shellView) statusReset() {
	s.logo().reset()
	s.Draw()
}

// View Accessors...

func (s *shellView) crumbs() *crumbsView {
	return s.views["crumbs"].(*crumbsView)
}

func (s *shellView) logo() *logoView {
	return s.views["logo"].(*logoView)
}

func (s *appView) clusterInfo() *clusterInfoView {
	return s.views["clusterInfo"].(*clusterInfoView)
}

func (s *appView) flash() *flashView {
	return s.views["flash"].(*flashView)
}

func (s *appView) cmd() *cmdView {
	return s.views["cmd"].(*cmdView)
}
