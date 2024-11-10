// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
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
type Container struct {
	Base
}

// ColorerFunc colors a resource row.
func (c Container) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)

		idx, ok := h.IndexOf("STATE", true)
		if !ok {
			return c
		}
		switch strings.TrimSpace(re.Row.Fields[idx]) {
		case Pending:
			return model1.PendingColor
		case ContainerCreating, PodInitializing:
			return model1.AddColor
		case Terminating, Initialized:
			return model1.HighlightColor
		case Completed:
			return model1.CompletedColor
		case Running:
			return c
		default:
			return model1.ErrColor
		}
	}
}

// Header returns a header row.
func (Container) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "IDX"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "PF"},
		model1.HeaderColumn{Name: "IMAGE"},
		model1.HeaderColumn{Name: "READY"},
		model1.HeaderColumn{Name: "STATE"},
		model1.HeaderColumn{Name: "RESTARTS", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "PROBES(L:R:S)"},
		model1.HeaderColumn{Name: "CPU", Align: tview.AlignRight, MX: true},
		model1.HeaderColumn{Name: "MEM", Align: tview.AlignRight, MX: true},
		model1.HeaderColumn{Name: "CPU/R:L", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "MEM/R:L", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "%CPU/R", Align: tview.AlignRight, MX: true},
		model1.HeaderColumn{Name: "%CPU/L", Align: tview.AlignRight, MX: true},
		model1.HeaderColumn{Name: "%MEM/R", Align: tview.AlignRight, MX: true},
		model1.HeaderColumn{Name: "%MEM/L", Align: tview.AlignRight, MX: true},
		model1.HeaderColumn{Name: "PORTS"},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (c Container) Render(o interface{}, name string, r *model1.Row) error {
	co, ok := o.(ContainerRes)
	if !ok {
		return fmt.Errorf("expected ContainerRes, but got %T", o)
	}

	cur, res := gatherMetrics(co.Container, co.MX)
	ready, state, restarts := "false", MissingValue, "0"
	if co.Status != nil {
		ready, state, restarts = boolToStr(co.Status.Ready), ToContainerState(co.Status.State), strconv.Itoa(int(co.Status.RestartCount))
	}

	r.ID = co.Container.Name
	r.Fields = model1.Fields{
		co.Idx,
		co.Container.Name,
		"●",
		co.Container.Image,
		ready,
		state,
		restarts,
		probe(co.Container.LivenessProbe) + ":" + probe(co.Container.ReadinessProbe) + ":" + probe(co.Container.StartupProbe),
		toMc(cur.cpu),
		toMi(cur.mem),
		toMc(res.cpu) + ":" + toMc(res.lcpu),
		toMi(res.mem) + ":" + toMi(res.lmem),
		client.ToPercentageStr(cur.cpu, res.cpu),
		client.ToPercentageStr(cur.cpu, res.lcpu),
		client.ToPercentageStr(cur.mem, res.mem),
		client.ToPercentageStr(cur.mem, res.lmem),
		ToContainerPorts(co.Container.Ports),
		AsStatus(c.diagnose(state, ready)),
		ToAge(co.Age),
	}

	return nil
}

// Happy returns true if resource is happy, false otherwise.
func (Container) diagnose(state, ready string) error {
	if state == "Completed" {
		return nil
	}

	if ready == "false" {
		return errors.New("container is not ready")
	}
	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func containerRequests(co *v1.Container) v1.ResourceList {
	req := co.Resources.Requests
	if len(req) != 0 {
		return req
	}
	lim := co.Resources.Limits
	if len(lim) != 0 {
		return lim
	}

	return nil
}

func gatherMetrics(co *v1.Container, mx *mv1beta1.ContainerMetrics) (c, r metric) {
	rList, lList := containerRequests(co), co.Resources.Limits
	if rList.Cpu() != nil {
		r.cpu = rList.Cpu().MilliValue()
	}
	if rList.Memory() != nil {
		r.mem = rList.Memory().Value()
	}
	if lList.Cpu() != nil {
		r.lcpu = lList.Cpu().MilliValue()
	}
	if lList.Memory() != nil {
		r.lmem = lList.Memory().Value()
	}
	if mx != nil {
		if mx.Usage.Cpu() != nil {
			c.cpu = mx.Usage.Cpu().MilliValue()
		}
		if mx.Usage.Memory() != nil {
			c.mem = mx.Usage.Memory().Value()
		}
	}

	return
}

// ToContainerPorts returns container ports as a string.
func ToContainerPorts(pp []v1.ContainerPort) string {
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

// ToContainerState returns container state as a string.
func ToContainerState(s v1.ContainerState) string {
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
		return "Terminating"
	case s.Running != nil:
		return "Running"
	default:
		return MissingValue
	}
}

const (
	on  = "on"
	off = "off"
)

func probe(p *v1.Probe) string {
	if p == nil {
		return off
	}
	return on
}

// ContainerRes represents a container and its metrics.
type ContainerRes struct {
	Container *v1.Container
	Status    *v1.ContainerStatus
	MX        *mv1beta1.ContainerMetrics
	Idx       string
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
