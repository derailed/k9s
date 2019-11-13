package view

import (
	"context"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
)

type PageStack struct {
	*ui.Pages

	app *App
}

func NewPageStack() *PageStack {
	return &PageStack{
		Pages: ui.NewPages(),
	}
}

func (p *PageStack) Init(ctx context.Context) {
	p.app = ctx.Value(ui.KeyApp).(*App)

	p.Pages.SetChangedFunc(func() {
		log.Debug().Msgf(">>>>>PS CHNGED<<<<<")
		p.DumpStack()
		active := p.CurrentPage()
		if active == nil {
			return
		}
		c := active.Item.(model.Component)
		log.Debug().Msgf("-------Page activated %#v", active)
		p.app.Hint.SetHints(c.Hints())
	})

	p.Pages.SetTitle("Fuck!")
	p.Stack.AddListener(p)
}

func (p *PageStack) StackPushed(c model.Component) {
	ctx := context.WithValue(context.Background(), ui.KeyApp, p.app)
	c.Init(ctx)
	p.app.SetFocus(c)
	p.app.Hint.SetHints(c.Hints())
}

func (p *PageStack) StackPopped(o, top model.Component) {
	o.Stop()
	p.StackTop(top)
}

func (p *PageStack) StackTop(top model.Component) {
	if top == nil {
		return
	}
	top.Start()
	p.app.SetFocus(top)
	p.app.Hint.SetHints(top.Hints())
}
