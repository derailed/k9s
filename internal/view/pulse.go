package view

import (
	"context"
	"fmt"
	"image"
	"log/slog"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/tchart"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	cpuFmt     = " %s [%s::b]%s[white::-]([%s::]%sm[white::]/[%s::]%sm[-::])"
	memFmt     = " %s [%s::b]%s[white::-]([%s::]%sMi[white::]/[%s::]%sMi[-::])"
	pulseTitle = "Pulses"
	NSTitleFmt = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-] "
	dirLeft    = 1
	dirRight   = -dirLeft
	dirDown    = 4
	dirUp      = -dirDown
	grayC      = "gray"
)

var corpusGVRs = append(model.PulseGVRs, client.CpuGVR, client.MemGVR)

type Charts map[*client.GVR]Graphable

// Graphable represents a graphic component.
type Graphable interface {
	tview.Primitive

	// ID returns the graph id.
	ID() string

	// Add adds a metric
	Add(ok, fault int)

	AddMetric(time.Time, float64)

	// SetLegend sets the graph legend
	SetLegend(string)

	SetColorIndex(int)

	SetMax(float64)
	GetMax() float64

	// SetSeriesColors sets charts series colors.
	SetSeriesColors(...tcell.Color)

	// GetSeriesColorNames returns the series color names.
	GetSeriesColorNames() []string

	// SetFocusColorNames sets the focus color names.
	SetFocusColorNames(fg, bg string)

	// SetBackgroundColor sets chart bg color.
	SetBackgroundColor(tcell.Color)

	SetBorderColor(tcell.Color) *tview.Box

	// IsDial returns true if chart is a dial
	IsDial() bool
}

// Pulse represents a command health view.
type Pulse struct {
	*tview.Grid

	app            *App
	gvr            *client.GVR
	model          *model.Pulse
	cancelFn       context.CancelFunc
	actions        *ui.KeyActions
	charts         Charts
	prevFocusIndex int
	chartGVRs      client.GVRs
}

// NewPulse returns a new alias view.
func NewPulse(gvr *client.GVR) ResourceViewer {
	return &Pulse{
		Grid:           tview.NewGrid(),
		model:          model.NewPulse(gvr),
		actions:        ui.NewKeyActions(),
		prevFocusIndex: -1,
	}
}

// Init initializes the view.
func (p *Pulse) Init(ctx context.Context) error {
	p.SetBorder(true)
	p.SetGap(0, 0)
	p.SetBorderPadding(0, 0, 1, 1)
	var err error
	if p.app, err = extractApp(ctx); err != nil {
		return err
	}

	ns := p.app.Config.ActiveNamespace()
	frame := p.app.Styles.Frame()
	p.SetTitle(ui.SkinTitle(fmt.Sprintf(NSTitleFmt, pulseTitle, ns), &frame))

	index, chartRow := 4, 6
	if client.IsAllNamespace(ns) {
		index, chartRow = 0, 8
	}
	p.chartGVRs = corpusGVRs[index:]

	p.charts = make(Charts, len(p.chartGVRs))
	var x, y, col int
	for _, gvr := range p.chartGVRs[:len(p.chartGVRs)-2] {
		p.charts[gvr] = p.makeGA(image.Point{X: x, Y: y}, image.Point{X: 2, Y: 2}, gvr)
		col, y = col+1, y+2
		if y > 6 {
			y = 0
		}
		if col >= 4 {
			col, x = 0, x+2
		}
	}
	if p.app.Conn().HasMetrics() {
		p.charts[client.CpuGVR] = p.makeSP(image.Point{X: chartRow, Y: 0}, image.Point{X: 2, Y: 4}, client.CpuGVR, "c")
		p.charts[client.MemGVR] = p.makeSP(image.Point{X: chartRow, Y: 4}, image.Point{X: 2, Y: 4}, client.MemGVR, "Gi")
	}
	p.GetItem(0).Focus = true
	p.app.SetFocus(p.charts[p.chartGVRs[0]])

	p.bindKeys()
	p.app.Styles.AddListener(p)
	p.StylesChanged(p.app.Styles)
	p.model.SetNamespace(ns)

	return nil
}

// InCmdMode checks if prompt is active.
func (*Pulse) InCmdMode() bool {
	return false
}

func (*Pulse) SetCommand(*cmd.Interpreter)            {}
func (*Pulse) SetFilter(string, bool)                 {}
func (*Pulse) SetLabelSelector(labels.Selector, bool) {}

// StylesChanged notifies the skin changed.
func (p *Pulse) StylesChanged(s *config.Styles) {
	p.SetBackgroundColor(s.Charts().BgColor.Color())
	for _, c := range p.charts {
		c.SetFocusColorNames(s.Charts().FocusFgColor.String(), s.Charts().FocusBgColor.String())
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

// SeriesChanged update cluster time series.
func (p *Pulse) SeriesChanged(tt dao.TimeSeries) {
	if len(tt) == 0 {
		return
	}

	cpu, ok := p.charts[client.CpuGVR]
	if !ok {
		return
	}
	mem := p.charts[client.MemGVR]
	if !ok {
		return
	}

	for i := range tt {
		t := tt[i]
		cpu.SetMax(float64(t.Value.AllocatableCPU))
		mem.SetMax(float64(t.Value.AllocatableMEM))
		cpu.AddMetric(t.Time, float64(t.Value.CurrentCPU))
		mem.AddMetric(t.Time, float64(t.Value.CurrentMEM))
	}

	last := tt[len(tt)-1]
	perc := client.ToPercentage(last.Value.CurrentCPU, int64(cpu.GetMax()))
	index := int(p.app.Config.K9s.Thresholds.LevelFor("cpu", perc))
	cpu.SetColorIndex(int(p.app.Config.K9s.Thresholds.LevelFor("cpu", perc)))
	nn := cpu.GetSeriesColorNames()
	if last.Value.CurrentCPU == 0 {
		nn[0] = grayC
	}
	if last.Value.AllocatableCPU == 0 {
		nn[1] = grayC
	}
	cpu.SetLegend(fmt.Sprintf(cpuFmt,
		cases.Title(language.English).String(client.CpuGVR.R()),
		p.app.Config.K9s.Thresholds.SeverityColor("cpu", perc),
		render.PrintPerc(perc),
		nn[index],
		render.AsThousands(last.Value.CurrentCPU),
		"white",
		render.AsThousands(int64(cpu.GetMax())),
	))

	nn = mem.GetSeriesColorNames()
	if last.Value.CurrentMEM == 0 {
		nn[0] = grayC
	}
	if last.Value.AllocatableMEM == 0 {
		nn[1] = grayC
	}
	perc = client.ToPercentage(last.Value.CurrentMEM, int64(mem.GetMax()))
	index = int(p.app.Config.K9s.Thresholds.LevelFor("memory", perc))
	mem.SetColorIndex(index)
	mem.SetLegend(fmt.Sprintf(memFmt,
		cases.Title(language.English).String(client.MemGVR.R()),
		p.app.Config.K9s.Thresholds.SeverityColor("memory", perc),
		render.PrintPerc(perc),
		nn[index],
		render.AsThousands(last.Value.CurrentMEM),
		"white",
		render.AsThousands(int64(mem.GetMax())),
	))
}

// PulseChanged notifies the model data changed.
func (p *Pulse) PulseChanged(pt model.HealthPoint) {
	v, ok := p.charts[pt.GVR]
	if !ok {
		return
	}

	nn := v.GetSeriesColorNames()
	if pt.Total == 0 {
		nn[0] = grayC
	}
	if pt.Faults == 0 {
		nn[1] = grayC
	}

	v.SetLegend(cases.Title(language.English).String(pt.GVR.R()))
	if pt.Faults > 0 {
		v.SetBorderColor(tcell.ColorDarkRed)
	} else {
		v.SetBorderColor(tcell.ColorDarkOliveGreen)
	}
	v.Add(pt.Total, pt.Faults)
}

// PulseFailed notifies the load failed.
func (p *Pulse) PulseFailed(err error) {
	p.app.Flash().Err(err)
}

func (p *Pulse) bindKeys() {
	p.actions.Merge(ui.NewKeyActionsFromMap(ui.KeyMap{
		tcell.KeyEnter:   ui.NewKeyAction("Goto", p.enterCmd, true),
		tcell.KeyTab:     ui.NewKeyAction("Next", p.nextFocusCmd(dirLeft), true),
		tcell.KeyBacktab: ui.NewKeyAction("Prev", p.nextFocusCmd(dirRight), true),
		tcell.KeyDown:    ui.NewKeyAction("Next", p.nextFocusCmd(dirDown), false),
		tcell.KeyUp:      ui.NewKeyAction("Prev", p.nextFocusCmd(dirUp), false),
		tcell.KeyRight:   ui.NewKeyAction("Next", p.nextFocusCmd(dirLeft), false),
		tcell.KeyLeft:    ui.NewKeyAction("Next", p.nextFocusCmd(dirRight), false),
	}))
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

func (*Pulse) Restart() {}

// Start initializes resource watch loop.
func (p *Pulse) Start() {
	p.Stop()

	ctx := p.defaultContext()
	ctx, p.cancelFn = context.WithCancel(ctx)
	gaugeChan, metricsChan, err := p.model.Watch(ctx)
	if err != nil {
		slog.Error("Pulse watch failed", slogs.Error, err)
		return
	}

	go func() {
		for {
			select {
			case check, ok := <-gaugeChan:
				if !ok {
					return
				}
				p.app.QueueUpdateDraw(func() {
					p.PulseChanged(check)
				})
			case mx, ok := <-metricsChan:
				if !ok {
					return
				}
				p.app.QueueUpdateDraw(func() {
					p.SeriesChanged(mx)
				})
			}
		}
	}()
}

// Stop terminates watch loop.
func (p *Pulse) Stop() {
	if p.cancelFn == nil {
		return
	}
	p.cancelFn()
	p.cancelFn = nil
}

// Refresh updates the view
func (*Pulse) Refresh() {}

// GVR returns a resource descriptor.
func (p *Pulse) GVR() *client.GVR {
	return p.gvr
}

// Name returns the component name.
func (*Pulse) Name() string {
	return pulseTitle
}

// App returns the current app handle.
func (p *Pulse) App() *App {
	return p.app
}

// SetInstance sets specific resource instance.
func (*Pulse) SetInstance(string) {}

// SetEnvFn sets the custom environment function.
func (*Pulse) SetEnvFn(EnvFunc) {}

// AddBindKeysFn sets up extra key bindings.
func (*Pulse) AddBindKeysFn(BindKeysFunc) {}

// SetContextFn sets custom context.
func (*Pulse) SetContextFn(ContextFunc) {}

func (*Pulse) GetContextFn() ContextFunc { return nil }

// GetTable return the view table if any.
func (*Pulse) GetTable() *Table {
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
func (*Pulse) ExtraHints() map[string]string {
	return nil
}

func (p *Pulse) enterCmd(*tcell.EventKey) *tcell.EventKey {
	v := p.App().GetFocus()
	s, ok := v.(Graphable)
	if !ok {
		return nil
	}
	g, ok := v.(Graphable)
	if !ok {
		return nil
	}
	p.prevFocusIndex = p.findIndex(g)
	for i := range len(p.charts) {
		gi := p.GetItem(i)
		if i == p.prevFocusIndex {
			gi.Focus = true
		} else {
			gi.Focus = false
		}
	}

	p.Stop()
	res := client.NewGVR(s.ID()).R()
	if res == "cpu" || res == "memory" {
		res = client.PodGVR.String()
	}
	p.App().SetFocus(p.App().Main)
	p.App().gotoResource(res+" "+p.model.GetNamespace(), "", false, true)

	return nil
}

func (p *Pulse) nextFocusCmd(direction int) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(*tcell.EventKey) *tcell.EventKey {
		v := p.app.GetFocus()
		g, ok := v.(Graphable)
		if !ok {
			return nil
		}

		currentIndex := p.findIndex(g)
		nextIndex, total := currentIndex+direction, len(p.charts)
		if nextIndex < 0 {
			return nil
		}

		switch direction {
		case dirLeft:
			if nextIndex >= total {
				return nil
			}
			p.prevFocusIndex = -1
		case dirRight:
			p.prevFocusIndex = -1
		case dirUp:
			if p.app.Conn().HasMetrics() {
				if currentIndex >= total-2 {
					if p.prevFocusIndex >= 0 && p.prevFocusIndex != currentIndex {
						nextIndex = p.prevFocusIndex
					} else if currentIndex == p.chartGVRs.Len()-1 {
						nextIndex += 1
					}
				} else {
					p.prevFocusIndex = currentIndex
				}
			}
		case dirDown:
			if p.app.Conn().HasMetrics() {
				if currentIndex >= total-6 && currentIndex < total-2 {
					switch {
					case (currentIndex % 4) <= 1:
						p.prevFocusIndex, nextIndex = currentIndex, total-2
					case (currentIndex % 4) <= 3:
						p.prevFocusIndex, nextIndex = currentIndex, total-1
					}
				} else if currentIndex >= total-2 {
					return nil
				}
			}
		}
		if nextIndex < 0 {
			nextIndex = 0
		} else if nextIndex > total-1 {
			nextIndex = currentIndex
		}
		p.GetItem(nextIndex).Focus = false
		p.GetItem(nextIndex).Item.Blur()
		i, v := p.nextFocus(nextIndex)
		p.GetItem(i).Focus = true
		p.app.SetFocus(v)

		return nil
	}
}

func (p *Pulse) makeSP(loc, span image.Point, gvr *client.GVR, unit string) *tchart.SparkLine {
	s := tchart.NewSparkLine(gvr.String(), unit)
	s.SetBackgroundColor(p.app.Styles.Charts().BgColor.Color())
	if cc, ok := p.app.Styles.Charts().ResourceColors[gvr.String()]; ok {
		s.SetSeriesColors(cc.Colors()...)
	} else {
		s.SetSeriesColors(p.app.Styles.Charts().DefaultChartColors.Colors()...)
	}
	s.SetLegend(fmt.Sprintf(" %s ", cases.Title(language.English).String(gvr.R())))
	s.SetInputCapture(p.keyboard)
	p.AddItem(s, loc.X, loc.Y, span.X, span.Y, 0, 0, false)

	return s
}

func (p *Pulse) makeGA(loc, span image.Point, gvr *client.GVR) *tchart.Gauge {
	g := tchart.NewGauge(gvr.String())
	g.SetBorder(true)
	g.SetBackgroundColor(p.app.Styles.Charts().BgColor.Color())
	if cc, ok := p.app.Styles.Charts().ResourceColors[gvr.String()]; ok {
		g.SetSeriesColors(cc.Colors()...)
	} else {
		g.SetSeriesColors(p.app.Styles.Charts().DefaultDialColors.Colors()...)
	}
	g.SetLegend(fmt.Sprintf(" %s ", cases.Title(language.English).String(gvr.R())))
	g.SetInputCapture(p.keyboard)
	p.AddItem(g, loc.X, loc.Y, span.X, span.Y, 0, 0, false)

	return g
}

// ----------------------------------------------------------------------------
// Helpers

func (p *Pulse) nextFocus(index int) (int, tview.Primitive) {
	if index >= len(p.chartGVRs) {
		return 0, p.charts[p.chartGVRs[0]]
	}

	if index < 0 {
		return len(p.chartGVRs) - 1, p.charts[p.chartGVRs[len(p.chartGVRs)-1]]
	}

	return index, p.charts[p.chartGVRs[index]]
}

func (p *Pulse) findIndex(g Graphable) int {
	for i, gvr := range p.chartGVRs {
		if gvr.String() == g.ID() {
			return i
		}
	}
	return 0
}
