// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"image"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/health"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/tchart"
	"github.com/derailed/k9s/internal/ui"
)

// Graphable represents a graphic component.
type Graphable interface {
	tview.Primitive

	// ID returns the graph id.
	ID() string

	// Add adds a metric
	Add(tchart.Metric)

	// SetLegend sets the graph legend
	SetLegend(string)

	// SetSeriesColors sets charts series colors.
	SetSeriesColors(...tcell.Color)

	// GetSeriesColorNames returns the series color names.
	GetSeriesColorNames() []string

	// SetFocusColorNames sets the focus color names.
	SetFocusColorNames(fg, bg string)

	// SetBackgroundColor sets chart bg color.
	SetBackgroundColor(tcell.Color)

	// IsDial returns true if chart is a dial
	IsDial() bool
}

const pulseTitle = "Pulses"

var _ ResourceViewer = (*Pulse)(nil)

// Pulse represents a command health view.
type Pulse struct {
	*tview.Grid

	app      *App
	gvr      client.GVR
	model    *model.Pulse
	cancelFn context.CancelFunc
	actions  *ui.KeyActions
	charts   []Graphable
}

// NewPulse returns a new alias view.
func NewPulse(gvr client.GVR) ResourceViewer {
	return &Pulse{
		Grid:    tview.NewGrid(),
		model:   model.NewPulse(gvr.String()),
		actions: ui.NewKeyActions(),
	}
}

func (p *Pulse) SetFilter(string)                 {}
func (p *Pulse) SetLabelFilter(map[string]string) {}

// Init initializes the view.
func (p *Pulse) Init(ctx context.Context) error {
	p.SetBorder(true)
	p.SetTitle(fmt.Sprintf(" %s ", pulseTitle))
	p.SetGap(1, 1)
	p.SetBorderPadding(0, 0, 1, 1)
	var err error
	if p.app, err = extractApp(ctx); err != nil {
		return err
	}

	p.charts = []Graphable{
		p.makeGA(image.Point{X: 0, Y: 0}, image.Point{X: 2, Y: 2}, "apps/v1/deployments"),
		p.makeGA(image.Point{X: 0, Y: 2}, image.Point{X: 2, Y: 2}, "apps/v1/replicasets"),
		p.makeGA(image.Point{X: 0, Y: 4}, image.Point{X: 2, Y: 2}, "apps/v1/statefulsets"),
		p.makeGA(image.Point{X: 0, Y: 6}, image.Point{X: 2, Y: 2}, "apps/v1/daemonsets"),
		p.makeSP(image.Point{X: 2, Y: 0}, image.Point{X: 3, Y: 2}, "v1/pods"),
		p.makeSP(image.Point{X: 2, Y: 2}, image.Point{X: 3, Y: 2}, "v1/events"),
		p.makeSP(image.Point{X: 2, Y: 4}, image.Point{X: 3, Y: 2}, "batch/v1/jobs"),
		p.makeSP(image.Point{X: 2, Y: 6}, image.Point{X: 3, Y: 2}, "v1/persistentvolumes"),
	}
	if p.app.Conn().HasMetrics() {
		p.charts = append(p.charts,
			p.makeSP(image.Point{X: 5, Y: 0}, image.Point{X: 2, Y: 4}, "cpu"),
			p.makeSP(image.Point{X: 5, Y: 4}, image.Point{X: 2, Y: 4}, "mem"),
		)
	}
	p.bindKeys()
	p.model.AddListener(p)
	p.app.SetFocus(p.charts[0])
	p.app.Styles.AddListener(p)
	p.StylesChanged(p.app.Styles)

	return nil
}

// InCmdMode checks if prompt is active.
func (*Pulse) InCmdMode() bool {
	return false
}

// StylesChanged notifies the skin changed.
func (p *Pulse) StylesChanged(s *config.Styles) {
	p.SetBackgroundColor(s.Charts().BgColor.Color())
	for _, c := range p.charts {
		c.SetFocusColorNames(s.Table().BgColor.String(), s.Table().CursorBgColor.String())
		if c.IsDial() {
			c.SetBackgroundColor(s.Charts().DialBgColor.Color())
			c.SetSeriesColors(s.Charts().DefaultDialColors.Colors()...)
		} else {
			c.SetBackgroundColor(s.Charts().ChartBgColor.Color())
			c.SetSeriesColors(s.Charts().DefaultChartColors.Colors()...)
		}
		if ss, ok := s.Charts().ResourceColors[c.ID()]; ok {
			c.SetSeriesColors(ss.Colors()...)
		}
	}
}

const (
	genFmat = " %s([%s::]%d[white::]:[%s::b]%d[-::])"
	cpuFmt  = " %s [%s::b]%s[white::-]([%s::]%sm[white::]/[%s::]%sm[-::])"
	memFmt  = " %s [%s::b]%s[white::-]([%s::]%sMi[white::]/[%s::]%sMi[-::])"
)

// PulseChanged notifies the model data changed.
func (p *Pulse) PulseChanged(c *health.Check) {
	index, ok := findIndexGVR(p.charts, c.GVR)
	if !ok {
		return
	}

	v, ok := p.GetItem(index).Item.(Graphable)
	if !ok {
		return
	}

	nn := v.GetSeriesColorNames()
	if c.Tally(health.S1) == 0 {
		nn[0] = "gray"
	}
	if c.Tally(health.S2) == 0 {
		nn[1] = "gray"
	}

	gvr := client.NewGVR(c.GVR)
	switch c.GVR {
	case "cpu":
		perc := client.ToPercentage(c.Tally(health.S1), c.Tally(health.S2))
		v.SetLegend(fmt.Sprintf(cpuFmt,
			cases.Title(language.Und, cases.NoLower).String(gvr.R()),
			p.app.Config.K9s.Thresholds.SeverityColor("cpu", perc),
			render.PrintPerc(perc),
			nn[0],
			render.AsThousands(c.Tally(health.S1)),
			nn[1],
			render.AsThousands(c.Tally(health.S2)),
		))
	case "mem":
		perc := client.ToPercentage(c.Tally(health.S1), c.Tally(health.S2))
		v.SetLegend(fmt.Sprintf(memFmt,
			cases.Title(language.Und, cases.NoLower).String(gvr.R()),
			p.app.Config.K9s.Thresholds.SeverityColor("memory", perc),
			render.PrintPerc(perc),
			nn[0],
			render.AsThousands(c.Tally(health.S1)),
			nn[1],
			render.AsThousands(c.Tally(health.S2)),
		))
	default:
		v.SetLegend(fmt.Sprintf(genFmat,
			cases.Title(language.Und, cases.NoLower).String(gvr.R()),
			nn[0],
			c.Tally(health.S1),
			nn[1],
			c.Tally(health.S2),
		))
	}
	v.Add(tchart.Metric{S1: c.Tally(health.S1), S2: c.Tally(health.S2)})
}

// PulseFailed notifies the load failed.
func (p *Pulse) PulseFailed(err error) {
	p.app.Flash().Err(err)
}

func (p *Pulse) bindKeys() {
	p.actions.Merge(ui.NewKeyActionsFromMap(ui.KeyMap{
		tcell.KeyEnter:   ui.NewKeyAction("Goto", p.enterCmd, true),
		tcell.KeyTab:     ui.NewKeyAction("Next", p.nextFocusCmd(1), true),
		tcell.KeyBacktab: ui.NewKeyAction("Prev", p.nextFocusCmd(-1), true),
	}))

	for i, v := range p.charts {
		t := cases.Title(language.Und, cases.NoLower).String(client.NewGVR(v.ID()).R())
		p.actions.Add(ui.NumKeys[i], ui.NewKeyAction(t, p.sparkFocusCmd(i), true))
	}
}

func (p *Pulse) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		key = tcell.Key(evt.Rune())
	}
	if a, ok := p.actions.Get(key); ok {
		return a.Action(evt)
	}

	return evt
}

func (p *Pulse) defaultContext() context.Context {
	return context.WithValue(context.Background(), internal.KeyFactory, p.app.factory)
}

// Start initializes resource watch loop.
func (p *Pulse) Start() {
	p.Stop()

	ctx := p.defaultContext()
	ctx, p.cancelFn = context.WithCancel(ctx)
	p.model.Watch(ctx)
}

// Stop terminates watch loop.
func (p *Pulse) Stop() {
	if p.cancelFn == nil {
		return
	}
	p.cancelFn()
	p.cancelFn = nil
}

// Refresh updates the view.
func (p *Pulse) Refresh() {}

// GVR returns a resource descriptor.
func (p *Pulse) GVR() client.GVR {
	return p.gvr
}

// Name returns the component name.
func (p *Pulse) Name() string {
	return pulseTitle
}

// App returns the current app handle.
func (p *Pulse) App() *App {
	return p.app
}

// SetInstance sets specific resource instance.
func (p *Pulse) SetInstance(string) {}

// SetEnvFn sets the custom environment function.
func (p *Pulse) SetEnvFn(EnvFunc) {}

// AddBindKeysFn sets up extra key bindings.
func (p *Pulse) AddBindKeysFn(BindKeysFunc) {}

// SetContextFn sets custom context.
func (p *Pulse) SetContextFn(ContextFunc) {}

// GetTable return the view table if any.
func (p *Pulse) GetTable() *Table {
	return nil
}

// Actions returns active menu bindings.
func (p *Pulse) Actions() *ui.KeyActions {
	return p.actions
}

// Hints returns the view hints.
func (p *Pulse) Hints() model.MenuHints {
	return p.actions.Hints()
}

// ExtraHints returns additional hints.
func (p *Pulse) ExtraHints() map[string]string {
	return nil
}

func (p *Pulse) sparkFocusCmd(i int) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		p.app.SetFocus(p.charts[i])
		return nil
	}
}

func (p *Pulse) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	v := p.App().GetFocus()
	s, ok := v.(Graphable)
	if !ok {
		return nil
	}
	res := client.NewGVR(s.ID()).R()
	if res == "cpu" || res == "mem" {
		res = "pod"
	}
	p.App().gotoResource(res+" all", "", false)

	return nil
}

func (p *Pulse) nextFocusCmd(direction int) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		v := p.app.GetFocus()
		index := findIndex(p.charts, v)
		p.GetItem(index).Focus = false
		p.GetItem(index).Item.Blur()
		i, v := nextFocus(p.charts, index+direction)
		p.GetItem(i).Focus = true
		p.app.SetFocus(v)

		return nil
	}
}

func (p *Pulse) makeSP(loc image.Point, span image.Point, gvr string) *tchart.SparkLine {
	s := tchart.NewSparkLine(gvr)
	s.SetBackgroundColor(p.app.Styles.Charts().BgColor.Color())
	s.SetBorderPadding(0, 1, 0, 1)
	if cc, ok := p.app.Styles.Charts().ResourceColors[gvr]; ok {
		s.SetSeriesColors(cc.Colors()...)
	} else {
		s.SetSeriesColors(p.app.Styles.Charts().DefaultChartColors.Colors()...)
	}
	s.SetLegend(fmt.Sprintf(" %s ", cases.Title(language.Und, cases.NoLower).String(client.NewGVR(gvr).R())))
	s.SetInputCapture(p.keyboard)
	s.SetMultiSeries(true)
	p.AddItem(s, loc.X, loc.Y, span.X, span.Y, 0, 0, true)

	return s
}

func (p *Pulse) makeGA(loc image.Point, span image.Point, gvr string) *tchart.Gauge {
	g := tchart.NewGauge(gvr)
	// g.SetResolution(3)
	g.SetBackgroundColor(p.app.Styles.Charts().BgColor.Color())
	// g.SetBorderPadding(0, 1, 0, 1)
	if cc, ok := p.app.Styles.Charts().ResourceColors[gvr]; ok {
		g.SetSeriesColors(cc.Colors()...)
	} else {
		g.SetSeriesColors(p.app.Styles.Charts().DefaultDialColors.Colors()...)
	}
	g.SetLegend(fmt.Sprintf(" %s ", cases.Title(language.Und, cases.NoLower).String(client.NewGVR(gvr).R())))
	g.SetInputCapture(p.keyboard)
	p.AddItem(g, loc.X, loc.Y, span.X, span.Y, 0, 0, true)

	return g
}

// ----------------------------------------------------------------------------
// Helpers

func nextFocus(pp []Graphable, index int) (int, tview.Primitive) {
	if index >= len(pp) {
		return 0, pp[0]
	}

	if index < 0 {
		return len(pp) - 1, pp[len(pp)-1]
	}

	return index, pp[index]
}

func findIndex(pp []Graphable, p tview.Primitive) int {
	for i, v := range pp {
		if v == p {
			return i
		}
	}
	return 0
}

func findIndexGVR(pp []Graphable, gvr string) (int, bool) {
	for i, v := range pp {
		if v.ID() == gvr {
			return i, true
		}
	}
	return 0, false
}
