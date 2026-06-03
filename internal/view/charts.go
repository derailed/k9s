// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
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
	chartsTitle    = "Charts"
	chartsTitleFmt = "[fg:bg:b] Charts([hilite:bg:b]%s[fg:bg:-])[fg:bg:-] "
	chartsInterval = 10 * time.Second
)

// ContainerCharts renders CPU and memory usage sparklines for a single container.
type ContainerCharts struct {
	*tview.Flex

	app       *App
	gvr       *client.GVR
	path      string
	container string
	cancelFn  context.CancelFunc
	cpu       *tchart.SparkLine
	mem       *tchart.SparkLine
	actions   *ui.KeyActions
}

var _ model.Component = (*ContainerCharts)(nil)

// NewContainerCharts returns a new container metrics chart view.
func NewContainerCharts(gvr *client.GVR, path, container string) *ContainerCharts {
	return &ContainerCharts{
		Flex:      tview.NewFlex(),
		gvr:       gvr,
		path:      path,
		container: container,
		actions:   ui.NewKeyActions(),
	}
}

func (*ContainerCharts) SetCommand(*cmd.Interpreter)            {}
func (*ContainerCharts) SetFilter(string, bool)                 {}
func (*ContainerCharts) SetLabelSelector(labels.Selector, bool) {}
func (*ContainerCharts) InCmdMode() bool                        { return false }

// Name returns the component name.
func (*ContainerCharts) Name() string { return chartsTitle }

// Init initializes the view.
func (c *ContainerCharts) Init(ctx context.Context) error {
	var err error
	if c.app, err = extractApp(ctx); err != nil {
		return err
	}

	c.SetBorder(true)
	c.SetDirection(tview.FlexRow)

	frame := c.app.Styles.Frame()
	c.SetTitle(ui.SkinTitle(fmt.Sprintf(chartsTitleFmt, c.path+"/"+c.container), &frame))
	c.SetBackgroundColor(c.app.Styles.Charts().BgColor.Color())

	c.cpu = tchart.NewSparkLine(client.CpuGVR.String(), "m")
	c.mem = tchart.NewSparkLine(client.MemGVR.String(), "Mi")

	c.cpu.SetLegend(fmt.Sprintf(" %s ", cases.Title(language.English).String(client.CpuGVR.R())))
	c.mem.SetLegend(fmt.Sprintf(" %s ", cases.Title(language.English).String(client.MemGVR.R())))

	c.applyStyles(c.app.Styles.Charts())

	c.AddItem(c.cpu, 0, 1, false)
	c.AddItem(c.mem, 0, 1, false)

	c.bindKeys()
	c.app.Styles.AddListener(c)

	return nil
}

func (c *ContainerCharts) applyStyles(styles config.Charts) {
	for _, sp := range []*tchart.SparkLine{c.cpu, c.mem} {
		sp.SetBackgroundColor(styles.ChartBgColor.Color())
		sp.SetFocusColorNames(styles.FocusFgColor.String(), styles.FocusBgColor.String())
		if cc, ok := styles.ResourceColors[sp.ID()]; ok {
			sp.SetSeriesColors(cc.Colors()...)
		} else {
			sp.SetSeriesColors(styles.DefaultChartColors.Colors()...)
		}
	}
}

// StylesChanged responds to skin changes.
func (c *ContainerCharts) StylesChanged(s *config.Styles) {
	c.SetBackgroundColor(s.Charts().BgColor.Color())
	c.applyStyles(s.Charts())
}

func (c *ContainerCharts) bindKeys() {
	c.actions.Bulk(ui.KeyMap{
		tcell.KeyEscape: ui.NewKeyAction("Back", c.app.PrevCmd, false),
		ui.KeyQ:         ui.NewKeyAction("Back", c.app.PrevCmd, false),
	})
}

// Hints returns the menu hints.
func (c *ContainerCharts) Hints() model.MenuHints {
	return c.actions.Hints()
}

// ExtraHints returns additional hints.
func (*ContainerCharts) ExtraHints() map[string]string { return nil }

// Start begins polling container metrics.
func (c *ContainerCharts) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFn = cancel
	c.setLimits()
	go c.poll(ctx)
}

// Stop cancels the polling goroutine and removes style listener.
func (c *ContainerCharts) Stop() {
	if c.cancelFn != nil {
		c.cancelFn()
		c.cancelFn = nil
	}
	c.app.Styles.RemoveListener(c)
}

func (c *ContainerCharts) setLimits() {
	po, err := fetchPod(c.app.factory, c.path)
	if err != nil {
		return
	}
	co, err := locateContainer(c.container, po.Spec.Containers)
	if err != nil {
		return
	}
	if lim := co.Resources.Limits; lim != nil {
		if cpu := lim.Cpu(); cpu != nil && cpu.Value() > 0 {
			c.cpu.SetMax(float64(cpu.MilliValue()))
		}
		if mem := lim.Memory(); mem != nil && mem.Value() > 0 {
			c.mem.SetMax(float64(client.ToMB(mem.Value())))
		}
	}
}

func (c *ContainerCharts) poll(ctx context.Context) {
	c.fetchAndUpdate(ctx)
	ticker := time.NewTicker(chartsInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.fetchAndUpdate(ctx)
		}
	}
}

func (c *ContainerCharts) fetchAndUpdate(ctx context.Context) {
	dial := client.DialMetrics(c.app.Conn())
	cmx, err := dial.FetchContainersMetrics(ctx, c.path)
	if err != nil {
		slog.Warn("Fetch container metrics failed",
			slogs.Error, err,
			"path", c.path,
		)
		return
	}
	mx, ok := cmx[c.container]
	if !ok {
		return
	}

	now := time.Now()
	cpuVal := float64(mx.Usage.Cpu().MilliValue())
	memVal := float64(client.ToMB(mx.Usage.Memory().Value()))

	c.app.QueueUpdateDraw(func() {
		c.cpu.SetMax(cpuVal)
		c.mem.SetMax(memVal)
		c.cpu.AddMetric(now, cpuVal)
		c.mem.AddMetric(now, memVal)
		c.updateLegends(cpuVal, memVal)
	})
}

func (c *ContainerCharts) updateLegends(cpuVal, memVal float64) {
	nn := c.cpu.GetSeriesColorNames()
	perc := client.ToPercentage(int64(cpuVal), int64(c.cpu.GetMax()))
	idx := int(c.app.Config.K9s.Thresholds.LevelFor("cpu", perc))
	c.cpu.SetColorIndex(idx)
	c.cpu.SetLegend(fmt.Sprintf(cpuFmt,
		cases.Title(language.English).String(client.CpuGVR.R()),
		c.app.Config.K9s.Thresholds.SeverityColor("cpu", perc),
		render.PrintPerc(perc),
		nn[idx%len(nn)],
		render.AsThousands(int64(cpuVal)),
		"white",
		render.AsThousands(int64(c.cpu.GetMax())),
	))

	nn = c.mem.GetSeriesColorNames()
	perc = client.ToPercentage(int64(memVal), int64(c.mem.GetMax()))
	idx = int(c.app.Config.K9s.Thresholds.LevelFor("memory", perc))
	c.mem.SetColorIndex(idx)
	c.mem.SetLegend(fmt.Sprintf(memFmt,
		cases.Title(language.English).String(client.MemGVR.R()),
		c.app.Config.K9s.Thresholds.SeverityColor("memory", perc),
		render.PrintPerc(perc),
		nn[idx%len(nn)],
		render.AsThousands(int64(memVal)),
		"white",
		render.AsThousands(int64(c.mem.GetMax())),
	))
}
