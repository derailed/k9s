package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/tview"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// ContainerWithMetrics represents a container and it's metrics.
type ContainerWithMetrics interface {
	// Container returns the container
	Container() *v1.Container

	// ContainerStatus returns the current container status.
	ContainerStatus() *v1.ContainerStatus

	// Metrics returns the container metrics.
	Metrics() *mv1beta1.ContainerMetrics

	// Age returns the pod age.
	Age() metav1.Time

	// IsInit indicates a init container.
	IsInit() bool
}

// Container renders a K8s Container to screen.
type Container struct{}

// ColorerFunc colors a resource row.
func (Container) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Container) Header(ns string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "IMAGE"},
		Header{Name: "READY"},
		Header{Name: "STATE"},
		Header{Name: "INIT"},
		Header{Name: "RS", Align: tview.AlignRight},
		Header{Name: "PROBES(L:R)"},
		Header{Name: "CPU", Align: tview.AlignRight},
		Header{Name: "MEM", Align: tview.AlignRight},
		Header{Name: "%CPU", Align: tview.AlignRight},
		Header{Name: "%MEM", Align: tview.AlignRight},
		Header{Name: "PORTS"},
		Header{Name: "AGE"},
	}
}

// Render renders a K8s resource to screen.
func (Container) Render(o interface{}, name string, r *Row) error {
	oo, ok := o.(ContainerWithMetrics)
	if !ok {
		return fmt.Errorf("Expected ContainerWithMetrics, but got %T", o)
	}

	co, cs := oo.Container(), oo.ContainerStatus()

	c, p := gatherMetrics(co, oo.Metrics())
	ready, state, restarts := "false", MissingValue, "0"
	if cs != nil {
		ready, state, restarts = boolToStr(cs.Ready), toState(cs.State), strconv.Itoa(int(cs.RestartCount))
	}

	fields := make(Fields, 0, len(r.Fields))
	fields = append(fields,
		co.Name,
		co.Image,
		ready,
		state,
		boolToStr(oo.IsInit()),
		restarts,
		probe(co.LivenessProbe)+":"+probe(co.ReadinessProbe),
		c.cpu,
		c.mem,
		p.cpu,
		p.mem,
		toStrPorts(co.Ports),
		toAge(oo.Age()),
	)
	r.ID, r.Fields = co.Name, fields

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// func findContainer(po v1.Pod, n string) *v1.Container {
// 	for _, c := range po.Spec.InitContainers {
// 		if c.Name == n {
// 			return &c
// 		}
// 	}
// 	for _, c := range po.Spec.Containers {
// 		if c.Name == n {
// 			return &c
// 		}
// 	}

// 	return nil
// }

func gatherMetrics(co *v1.Container, mx *mv1beta1.ContainerMetrics) (c, p metric) {
	c, p = noMetric(), noMetric()
	if mx == nil {
		return
	}

	cpu := mx.Usage.Cpu().MilliValue()
	mem := k8s.ToMB(mx.Usage.Memory().Value())
	c = metric{
		cpu: ToMillicore(cpu),
		mem: ToMi(mem),
	}

	rcpu, rmem := containerResources(co)
	if rcpu != nil {
		p.cpu = AsPerc(toPerc(float64(cpu), float64(rcpu.MilliValue())))
	}
	if rmem != nil {
		p.mem = AsPerc(toPerc(mem, k8s.ToMB(rmem.Value())))
	}

	return
}

func toStrPorts(pp []v1.ContainerPort) string {
	ports := make([]string, len(pp))
	for i, p := range pp {
		if len(p.Name) > 0 {
			ports[i] = p.Name + ":"
		}
		ports[i] += strconv.Itoa(int(p.ContainerPort))
		if p.Protocol != "TCP" {
			ports[i] += "╱" + string(p.Protocol)
		}
	}

	return strings.Join(ports, ",")
}

func toState(s v1.ContainerState) string {
	switch {
	case s.Waiting != nil:
		if s.Waiting.Reason != "" {
			return s.Waiting.Reason
		}
		return "Waiting"

	case s.Terminated != nil:
		if s.Terminated.Reason != "" {
			return s.Terminated.Reason
		}
		return "Terminated"
	case s.Running != nil:
		return "Running"
	default:
		return MissingValue
	}
}

func toRes(r v1.ResourceList) (string, string) {
	cpu, mem := r[v1.ResourceCPU], r[v1.ResourceMemory]

	return ToMillicore(cpu.MilliValue()), ToMi(k8s.ToMB(mem.Value()))
}

func probe(p *v1.Probe) string {
	if p == nil {
		return "off"
	}
	return "on"
}

func asMi(v int64) float64 {
	return float64(v) / 1024 * 1024
}
