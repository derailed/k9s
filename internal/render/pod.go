package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/util/node"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// Pod renders a K8s Pod to screen.
type Pod struct{}

// ColorerFunc colors a resource row.
func (p Pod) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		c := DefaultColorer(ns, h, re)

		statusCol := h.IndexOf("STATUS", true)
		if statusCol == -1 {
			return c
		}
		status := strings.TrimSpace(re.Row.Fields[statusCol])
		switch status {
		case ContainerCreating, PodInitializing:
			c = AddColor
		case Initialized:
			c = HighlightColor
		case Completed:
			c = CompletedColor
		case Running:
			c = StdColor
			if !Happy(ns, h, re.Row) {
				c = ErrColor
			}
		case Terminating:
			c = KillColor
		default:
			if !Happy(ns, h, re.Row) {
				c = ErrColor
			}
		}
		return c
	}
}

// Header returns a header row.
func (Pod) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "READY"},
		HeaderColumn{Name: "RESTARTS", Align: tview.AlignRight},
		HeaderColumn{Name: "STATUS"},
		HeaderColumn{Name: "CPU", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "MEM", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "%CPU/R", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "%MEM/R", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "%CPU/L", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "%MEM/L", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "IP"},
		HeaderColumn{Name: "NODE"},
		HeaderColumn{Name: "QOS", Wide: true},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true, Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (p Pod) Render(o interface{}, ns string, r *Row) error {
	pwm, ok := o.(*PodWithMetrics)
	if !ok {
		return fmt.Errorf("Expected PodWithMetrics, but got %T", o)
	}

	var po v1.Pod
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(pwm.Raw.Object, &po)
	if err != nil {
		return err
	}

	ss := po.Status.ContainerStatuses
	cr, _, rc := p.Statuses(ss)
	c, perc := p.gatherPodMX(&po, pwm.MX)
	phase := p.Phase(&po)
	r.ID = client.MetaFQN(po.ObjectMeta)
	r.Fields = Fields{
		po.Namespace,
		po.ObjectMeta.Name,
		strconv.Itoa(cr) + "/" + strconv.Itoa(len(ss)),
		strconv.Itoa(rc),
		phase,
		c.cpu,
		c.mem,
		perc.cpu,
		perc.mem,
		perc.cpuLim,
		perc.memLim,
		na(po.Status.PodIP),
		na(po.Spec.NodeName),
		p.mapQOS(po.Status.QOSClass),
		mapToStr(po.Labels),
		asStatus(p.diagnose(phase, cr, len(ss))),
		toAge(po.ObjectMeta.CreationTimestamp),
	}

	return nil
}

func (p Pod) diagnose(phase string, cr, ct int) error {
	if phase == Completed {
		return nil
	}
	if cr != ct || ct == 0 {
		return fmt.Errorf("container ready check failed: %d of %d", cr, ct)
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// PodWithMetrics represents a pod and its metrics.
type PodWithMetrics struct {
	Raw *unstructured.Unstructured
	MX  *mv1beta1.PodMetrics
}

// GetObjectKind returns a schema object.
func (p *PodWithMetrics) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (p *PodWithMetrics) DeepCopyObject() runtime.Object {
	return p
}

func (*Pod) gatherPodMX(pod *v1.Pod, mx *mv1beta1.PodMetrics) (c, p metric) {
	c, p = noMetric(), noMetric()
	if mx == nil {
		return
	}

	cpu, mem := currentRes(mx)
	c = metric{
		cpu: ToMillicore(cpu.MilliValue()),
		mem: ToMi(client.ToMB(mem.Value())),
	}

	rc, rm := requestedRes(pod.Spec.Containers)
	lc, lm := resourceLimits(pod.Spec.Containers)
	p = metric{
		cpu:    IntToStr(client.ToPercentage(cpu.MilliValue(), rc.MilliValue())),
		mem:    IntToStr(client.ToPercentage(client.ToMB(mem.Value()), client.ToMB(rm.Value()))),
		cpuLim: IntToStr(client.ToPercentage(cpu.MilliValue(), lc.MilliValue())),
		memLim: IntToStr(client.ToPercentage(client.ToMB(mem.Value()), client.ToMB(lm.Value()))),
	}

	return
}

func containerResources(co v1.Container) (cpu, mem *resource.Quantity) {
	req, limit := co.Resources.Requests, co.Resources.Limits
	switch {
	case len(req) != 0:
		cpu, mem = req.Cpu(), req.Memory()
	case len(limit) != 0:
		cpu, mem = limit.Cpu(), limit.Memory()
	}

	return
}

func containerLimits(co v1.Container) (cpu, mem *resource.Quantity) {
	limit := co.Resources.Limits
	if len(limit) == 0 {
		return nil, nil
	}
	return limit.Cpu(), limit.Memory()
}

func resourceLimits(cc []v1.Container) (cpu, mem resource.Quantity) {
	for _, co := range cc {
		limit := co.Resources.Limits
		if len(limit) == 0 {
			continue
		}
		if limit.Cpu() != nil {
			cpu.Add(*limit.Cpu())
		}
		if limit.Memory() != nil {
			mem.Add(*limit.Memory())
		}
	}
	return
}

func requestedRes(cc []v1.Container) (cpu, mem resource.Quantity) {
	for _, co := range cc {
		c, m := containerResources(co)
		if c != nil {
			cpu.Add(*c)
		}
		if m != nil {
			mem.Add(*m)
		}
	}
	return
}

func currentRes(mx *mv1beta1.PodMetrics) (cpu, mem resource.Quantity) {
	if mx == nil {
		return
	}
	for _, co := range mx.Containers {
		c, m := co.Usage.Cpu(), co.Usage.Memory()
		cpu.Add(*c)
		mem.Add(*m)
	}
	return
}

func (*Pod) mapQOS(class v1.PodQOSClass) string {
	switch class {
	case v1.PodQOSGuaranteed:
		return "GA"
	case v1.PodQOSBurstable:
		return "BU"
	default:
		return "BE"
	}
}

// Statuses reports current pod container statuses.
func (*Pod) Statuses(ss []v1.ContainerStatus) (cr, ct, rc int) {
	for _, c := range ss {
		if c.State.Terminated != nil {
			ct++
		}
		if c.Ready {
			cr = cr + 1
		}
		rc += int(c.RestartCount)
	}

	return
}

// Phase reports the given pod phase.
func (p *Pod) Phase(po *v1.Pod) string {
	status := string(po.Status.Phase)
	if po.Status.Reason != "" {
		if po.DeletionTimestamp != nil && po.Status.Reason == node.NodeUnreachablePodReason {
			return "Unknown"
		}
		status = po.Status.Reason
	}

	status, ok := p.initContainerPhase(po.Status, len(po.Spec.InitContainers), status)
	if ok {
		return status
	}

	status, ok = p.containerPhase(po.Status, status)
	if ok && status == Completed {
		status = Running
	}
	if po.DeletionTimestamp == nil {
		return status
	}

	return Terminating
}

func (*Pod) containerPhase(st v1.PodStatus, status string) (string, bool) {
	var running bool
	for i := len(st.ContainerStatuses) - 1; i >= 0; i-- {
		cs := st.ContainerStatuses[i]
		switch {
		case cs.State.Waiting != nil && cs.State.Waiting.Reason != "":
			status = cs.State.Waiting.Reason
		case cs.State.Terminated != nil && cs.State.Terminated.Reason != "":
			status = cs.State.Terminated.Reason
		case cs.State.Terminated != nil:
			if cs.State.Terminated.Signal != 0 {
				status = "Signal:" + strconv.Itoa(int(cs.State.Terminated.Signal))
			} else {
				status = "ExitCode:" + strconv.Itoa(int(cs.State.Terminated.ExitCode))
			}
		case cs.Ready && cs.State.Running != nil:
			running = true
		}
	}

	return status, running
}

func (*Pod) initContainerPhase(st v1.PodStatus, initCount int, status string) (string, bool) {
	for i, cs := range st.InitContainerStatuses {
		s := checkContainerStatus(cs, i, initCount)
		if s == "" {
			continue
		}
		return s, true
	}

	return status, false
}

// ----------------------------------------------------------------------------
// Helpers..

func checkContainerStatus(cs v1.ContainerStatus, i, initCount int) string {
	switch {
	case cs.State.Terminated != nil:
		if cs.State.Terminated.ExitCode == 0 {
			return ""
		}
		if cs.State.Terminated.Reason != "" {
			return "Init:" + cs.State.Terminated.Reason
		}
		if cs.State.Terminated.Signal != 0 {
			return "Init:Signal:" + strconv.Itoa(int(cs.State.Terminated.Signal))
		}
		return "Init:ExitCode:" + strconv.Itoa(int(cs.State.Terminated.ExitCode))
	case cs.State.Waiting != nil && cs.State.Waiting.Reason != "" && cs.State.Waiting.Reason != "PodInitializing":
		return "Init:" + cs.State.Waiting.Reason
	default:
		return "Init:" + strconv.Itoa(i) + "/" + strconv.Itoa(initCount)
	}
}
