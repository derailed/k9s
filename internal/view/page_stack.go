package view

import (
	"context"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
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
	p.Stack.AddListener(p)
}

func (p *PageStack) StackPushed(c model.Component) {
	ctx := context.WithValue(context.Background(), ui.KeyApp, p.app)
	c.Init(ctx)
	c.Start()
	p.app.SetFocus(c)
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
}
