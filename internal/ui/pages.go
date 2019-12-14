package ui

import (
	"fmt"

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

func (p *Pages) Show(c model.Component) {
	p.SwitchToPage(componentID(c))
}

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

func (p *Pages) DumpPages() {
	log.Debug().Msgf("Dumping Pages %p", p)
	for i, c := range p.Stack.Peek() {
		log.Debug().Msgf("%d -- %s -- %#v", i, componentID(c), p.GetPrimitive(componentID(c)))
	}
}

// Stack Protocol...

func (p *Pages) StackPushed(c model.Component) {
	p.addAndShow(c)
}

func (p *Pages) StackPopped(o, top model.Component) {
	log.Debug().Msgf("UI STACK POPPED!!!")
	p.delete(o)
}

func (p *Pages) StackTop(top model.Component) {
	if top == nil {
		return
	}
	p.Show(top)
}

// Helpers...

func componentID(c model.Component) string {
	if c.Name() == "" {
		panic("Component has no name")
	}
	return fmt.Sprintf("%s-%p", c.Name(), c)
}
