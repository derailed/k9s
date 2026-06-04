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
	chartsInterval = 30 * time.Second
)

// ContainerCharts renders CPU and memory usage sparklines for a single container.
type ContainerCharts struct {
	*tview.Flex

	app       *App
	path      string
	container string
	cancelFn  context.CancelFunc
	cpu       *tchart.SparkLine
	mem       *tchart.SparkLine
	actions   *ui.KeyActions
	// cpuLimit and memLimit hold the container's configured resource limits
	// (milliCPU and MiB respectively). Zero means no limit is set.
	cpuLimit float64
	memLimit float64
}

var _ model.Component = (*ContainerCharts)(nil)

// NewContainerCharts returns a new container metrics chart view.
func NewContainerCharts(path, container string) *ContainerCharts {
	return &ContainerCharts{
		Flex:      tview.NewFlex(),
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

	c.AddItem(c.cpu, 0, 1, true)
	c.AddItem(c.mem, 0, 1, false)

	c.bindKeys()
	c.SetInputCapture(c.keyboard)
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

func (c *ContainerCharts) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		key = tcell.Key(evt.Rune())
	}
	if a, ok := c.actions.Get(key); ok {
		return a.Action(evt)
	}
	return evt
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

// setLimits reads the container's resource limits and stores them so that
// updateLegends can compute a meaningful percentage denominator.
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
			c.cpuLimit = float64(cpu.MilliValue())
			c.cpu.SetMax(c.cpuLimit)
		}
		if mem := lim.Memory(); mem != nil && mem.Value() > 0 {
			c.memLimit = float64(client.ToMB(mem.Value()))
			c.mem.SetMax(c.memLimit)
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
		// For unlimited containers, auto-scale the chart to the observed peak.
		// When a limit is set, SetMax is a no-op here because it only increases.
		c.cpu.SetMax(cpuVal)
		c.mem.SetMax(memVal)
		c.cpu.AddMetric(now, cpuVal)
		c.mem.AddMetric(now, memVal)
		c.updateLegends(cpuVal, memVal)
	})
}

func (c *ContainerCharts) updateLegends(cpuVal, memVal float64) {
	nn := c.cpu.GetSeriesColorNames()
	var (
		cpuPerc    int
		cpuPercStr string
		cpuColor   string
		cpuCeiling string
		cpuIdx     int
	)
	if c.cpuLimit > 0 {
		cpuPerc = client.ToPercentage(int64(cpuVal), int64(c.cpuLimit))
		cpuPercStr = render.PrintPerc(cpuPerc)
		cpuColor = c.app.Config.K9s.Thresholds.SeverityColor("cpu", cpuPerc)
		cpuIdx = int(c.app.Config.K9s.Thresholds.LevelFor("cpu", cpuPerc))
		cpuCeiling = render.AsThousands(int64(c.cpuLimit))
	} else {
		cpuPercStr = client.NA
		cpuColor = "white"
		cpuCeiling = render.AsThousands(int64(c.cpu.GetMax()))
	}
	c.cpu.SetColorIndex(cpuIdx)
	c.cpu.SetLegend(fmt.Sprintf(cpuFmt,
		cases.Title(language.English).String(client.CpuGVR.R()),
		cpuColor,
		cpuPercStr,
		nn[cpuIdx%len(nn)],
		render.AsThousands(int64(cpuVal)),
		"white",
		cpuCeiling,
	))

	nn = c.mem.GetSeriesColorNames()
	var (
		memPerc    int
		memPercStr string
		memColor   string
		memCeiling string
		memIdx     int
	)
	if c.memLimit > 0 {
		memPerc = client.ToPercentage(int64(memVal), int64(c.memLimit))
		memPercStr = render.PrintPerc(memPerc)
		memColor = c.app.Config.K9s.Thresholds.SeverityColor("memory", memPerc)
		memIdx = int(c.app.Config.K9s.Thresholds.LevelFor("memory", memPerc))
		memCeiling = render.AsThousands(int64(c.memLimit))
	} else {
		memPercStr = client.NA
		memColor = "white"
		memCeiling = render.AsThousands(int64(c.mem.GetMax()))
	}
	c.mem.SetColorIndex(memIdx)
	c.mem.SetLegend(fmt.Sprintf(memFmt,
		cases.Title(language.English).String(client.MemGVR.R()),
		memColor,
		memPercStr,
		nn[memIdx%len(nn)],
		render.AsThousands(int64(memVal)),
		"white",
		memCeiling,
	))
}
