package ui

import (
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
)

type Pages struct {
	*tview.Pages
	*model.Stack
}

func NewPages() *Pages {
	p := Pages{
		Pages: tview.NewPages(),
		Stack: model.NewStack(),
	}
	p.Stack.AddListener(&p)

	return &p
}

// Get fetch a page given its name.
func (p *Pages) get(n string) model.Component {
	if comp, ok := p.GetPrimitive(n).(model.Component); ok {
		return comp
	}

	return nil
}

// AddAndShow adds a new page and bring it to front.
func (p *Pages) addAndShow(c model.Component) {
	p.add(c)
	p.Show(c.Name())
}

// Add adds a new page.
func (p *Pages) add(c model.Component) {
	p.AddPage(c.Name(), c, true, true)
}

// Delete removes a page.
func (p *Pages) delete(c model.Component) {
	p.RemovePage(c.Name())
}

// Show brings a named page forward.
func (p *Pages) Show(n string) {
	p.SwitchToPage(n)
}

func (p *Pages) DumpPages() {
	log.Debug().Msgf("Dumping Pages %p", p)
	for i, n := range p.Stack.Flatten() {
		log.Debug().Msgf("%d -- %s -- %#v", i, n, p.GetPrimitive(n))
	}
}

// Stack Protocol...

func (p *Pages) StackPushed(c model.Component) {
	p.addAndShow(c)
}

func (p *Pages) StackPopped(o, top model.Component) {
	p.delete(o)
}

func (p *Pages) StackTop(top model.Component) {
	if top == nil {
		return
	}
	p.Show(top.Name())
}
