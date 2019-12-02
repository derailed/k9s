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

func (p *PageStack) Init(ctx context.Context) (err error) {
	if p.app, err = extractApp(ctx); err != nil {
		return err
	}

	p.Stack.AddListener(p)

	return nil
}

func (p *PageStack) StackPushed(c model.Component) {
	ctx := context.WithValue(context.Background(), ui.KeyApp, p.app)
	if err := c.Init(ctx); err != nil {
		log.Error().Err(err).Msgf("Component Init failed!")
		p.app.Flash().Err(err)
		return
	}
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
