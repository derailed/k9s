package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	return func(ns string, r RowEvent) tcell.Color {
		c := DefaultColorer(ns, r)

		readyCol := 2
		if strings.TrimSpace(r.Row.Fields[readyCol]) == "false" {
			c = ErrColor
		}

		stateCol := readyCol + 1
		switch strings.TrimSpace(r.Row.Fields[stateCol]) {
		case ContainerCreating, PodInitializing:
			return AddColor
		case Terminating, Initialized:
			return HighlightColor
		case Completed:
			return CompletedColor
		case Running:
		default:
			c = ErrColor
		}

		return c
	}
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
		Header{Name: "%MAX-CPU", Align: tview.AlignRight},
		Header{Name: "%MAX-MEM", Align: tview.AlignRight},
		Header{Name: "PORTS"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (c Container) Render(o interface{}, name string, r *Row) error {
	co, ok := o.(ContainerRes)
	if !ok {
		return fmt.Errorf("Expected ContainerRes, but got %T", o)
	}

	cur, perc, limit := gatherMetrics(co.Container, co.MX)
	ready, state, restarts := "false", MissingValue, "0"
	if co.Status != nil {
		ready, state, restarts = boolToStr(co.Status.Ready), toState(co.Status.State), strconv.Itoa(int(co.Status.RestartCount))
	}

	r.ID = co.Container.Name
	r.Fields = make(Fields, 0, len(c.Header(client.AllNamespaces)))
	r.Fields = append(r.Fields,
		co.Container.Name,
		co.Container.Image,
		ready,
		state,
		boolToStr(co.IsInit),
		restarts,
		probe(co.Container.LivenessProbe)+":"+probe(co.Container.ReadinessProbe),
		cur.cpu,
		cur.mem,
		perc.cpu,
		perc.mem,
		limit.cpu,
		limit.mem,
		toStrPorts(co.Container.Ports),
		toAge(co.Age),
	)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func gatherMetrics(co *v1.Container, mx *mv1beta1.ContainerMetrics) (c, p, l metric) {
	c, p, l = noMetric(), noMetric(), noMetric()
	if mx == nil {
		return
	}

	cpu := mx.Usage.Cpu().MilliValue()
	mem := ToMB(mx.Usage.Memory().Value())
	c = metric{
		cpu: ToMillicore(cpu),
		mem: ToMi(mem),
	}

	rcpu, rmem := containerResources(*co)
	if rcpu != nil {
		p.cpu = AsPerc(toPerc(float64(cpu), float64(rcpu.MilliValue())))
	}
	if rmem != nil {
		p.mem = AsPerc(toPerc(mem, ToMB(rmem.Value())))
	}

	lcpu, lmem := containerLimits(*co)
	if lcpu != nil {
		l.cpu = AsPerc(toPerc(float64(cpu), float64(lcpu.MilliValue())))
	}
	if lmem != nil {
		l.mem = AsPerc(toPerc(mem, ToMB(lmem.Value())))
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

func probe(p *v1.Probe) string {
	if p == nil {
		return "off"
	}
	return "on"
}

// ContainerRes represents a container and its metrics.
type ContainerRes struct {
	Container *v1.Container
	Status    *v1.ContainerStatus
	MX        *mv1beta1.ContainerMetrics
	IsInit    bool
	Age       metav1.Time
}

// GetObjectKind returns a schema object.
func (c ContainerRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (c ContainerRes) DeepCopyObject() runtime.Object {
	return c
}
