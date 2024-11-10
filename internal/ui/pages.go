// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
)

// Pages represents a stack of view pages.
type Pages struct {
	*tview.Pages
	*model.Stack
}

// NewPages return a new view.
func NewPages() *Pages {
	p := Pages{
		Pages: tview.NewPages(),
		Stack: model.NewStack(),
	}
	p.Stack.AddListener(&p)

	return &p
}

// IsTopDialog checks if front page is a dialog.
func (p *Pages) IsTopDialog() bool {
	_, pa := p.GetFrontPage()
	switch pa.(type) {
	case *tview.ModalForm, *ModalList:
		return true
	default:
		return false
	}
}

// Show displays a given page.
func (p *Pages) Show(c model.Component) {
	p.SwitchToPage(componentID(c))
}

// Current returns the current component.
func (p *Pages) Current() model.Component {
	c := p.CurrentPage()
	if c == nil {
		return nil
	}

	return c.Item.(model.Component)
}

// AddAndShow adds a new page and bring it to front.
func (p *Pages) addAndShow(c model.Component) {
	p.add(c)
	p.Show(c)
}

// Add adds a new page.
func (p *Pages) add(c model.Component) {
	p.AddPage(componentID(c), c, true, true)
}

// Delete removes a page.
func (p *Pages) delete(c model.Component) {
	p.RemovePage(componentID(c))
}

// Dump for debug.
func (p *Pages) Dump() {
	log.Debug().Msgf("Dumping Pages %p", p)
	for i, c := range p.Stack.Peek() {
		log.Debug().Msgf("%d -- %s -- %#v", i, componentID(c), p.GetPrimitive(componentID(c)))
	}
}

// Stack Protocol...

// StackPushed notifies a new component was pushed.
func (p *Pages) StackPushed(c model.Component) {
	p.addAndShow(c)
}

// StackPopped notifies a component was removed.
func (p *Pages) StackPopped(o, top model.Component) {
	p.delete(o)
}

// StackTop notifies a new component is at the top of the stack.
func (p *Pages) StackTop(top model.Component) {
	if top == nil {
		return
	}
	p.Show(top)
}

// Helpers...

func componentID(c model.Component) string {
	if c.Name() == "" {
		log.Error().Msg("Component has no name")
	}
	return fmt.Sprintf("%s-%p", c.Name(), c)
}
