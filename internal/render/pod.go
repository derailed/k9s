// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/derailed/k9s/internal/client"
)

const (
	// NodeUnreachablePodReason is reason and message set on a pod when its state
	// cannot be confirmed as kubelet is unresponsive on the node it is (was) running.
	NodeUnreachablePodReason = "NodeLost" // k8s.io/kubernetes/pkg/util/node.NodeUnreachablePodReason
)

const (
	PhaseTerminating       = "Terminating"
	PhaseInitialized       = "Initialized"
	PhaseRunning           = "Running"
	PhaseNotReady          = "NoReady"
	PhaseCompleted         = "Completed"
	PhaseContainerCreating = "ContainerCreating"
	PhasePodInitializing   = "PodInitializing"
	PhaseUnknown           = "Unknown"
	PhaseCrashLoop         = "CrashLoopBackOff"
	PhaseError             = "Error"
	PhaseImagePullBackOff  = "ImagePullBackOff"
	PhaseOOMKilled         = "OOMKilled"
)

// Pod renders a K8s Pod to screen.
type Pod struct {
	Base
}

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
		case Pending:
			c = PendingColor
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
		HeaderColumn{Name: "PF"},
		HeaderColumn{Name: "READY"},
		HeaderColumn{Name: "STATUS"},
		HeaderColumn{Name: "RESTARTS", Align: tview.AlignRight},
		HeaderColumn{Name: "IP"},
		HeaderColumn{Name: "NODE"},
		HeaderColumn{Name: "NOMINATED NODE", Wide: true},
		HeaderColumn{Name: "READINESS GATES", Wide: true},
		HeaderColumn{Name: "CPU", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "MEM", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "CPU/R:L", Align: tview.AlignRight, Wide: true},
		HeaderColumn{Name: "MEM/R:L", Align: tview.AlignRight, Wide: true},
		HeaderColumn{Name: "%CPU/R", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "%CPU/L", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "%MEM/R", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "%MEM/L", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "QOS", Wide: true},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (p Pod) Render(o interface{}, ns string, row *Row) error {
	pwm, ok := o.(*PodWithMetrics)
	if !ok {
		return fmt.Errorf("expected PodWithMetrics, but got %T", o)
	}

	var po v1.Pod
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(pwm.Raw.Object, &po); err != nil {
		return err
	}

	ss := po.Status.ContainerStatuses
	cr, _, rc := p.Statuses(ss)

	c, r := p.gatherPodMX(&po, pwm.MX)
	phase := p.Phase(&po)
	row.ID = client.MetaFQN(po.ObjectMeta)
	row.Fields = Fields{
		po.Namespace,
		po.ObjectMeta.Name,
		"●",
		strconv.Itoa(cr) + "/" + strconv.Itoa(len(po.Spec.Containers)),
		phase,
		strconv.Itoa(rc),
		na(po.Status.PodIP),
		na(po.Spec.NodeName),
		asNominated(po.Status.NominatedNodeName),
		asReadinessGate(po),
		toMc(c.cpu),
		toMi(c.mem),
		toMc(r.cpu) + ":" + toMc(r.lcpu),
		toMi(r.mem) + ":" + toMi(r.lmem),
		client.ToPercentageStr(c.cpu, r.cpu),
		client.ToPercentageStr(c.cpu, r.lcpu),
		client.ToPercentageStr(c.mem, r.mem),
		client.ToPercentageStr(c.mem, r.lmem),
		p.mapQOS(po.Status.QOSClass),
		mapToStr(po.Labels),
		asStatus(p.diagnose(phase, cr, len(ss))),
		toAge(po.GetCreationTimestamp()),
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

func asNominated(n string) string {
	if n == "" {
		return MissingValue
	}
	return n
}

func asReadinessGate(pod v1.Pod) string {
	if len(pod.Spec.ReadinessGates) == 0 {
		return MissingValue
	}

	trueConditions := 0
	for _, readinessGate := range pod.Spec.ReadinessGates {
		conditionType := readinessGate.ConditionType
		for _, condition := range pod.Status.Conditions {
			if condition.Type == conditionType {
				if condition.Status == "True" {
					trueConditions++
				}
				break
			}
		}
	}

	return strconv.Itoa(trueConditions) + "/" + strconv.Itoa(len(pod.Spec.ReadinessGates))
}

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

func (*Pod) gatherPodMX(pod *v1.Pod, mx *mv1beta1.PodMetrics) (c, r metric) {
	rcpu, rmem := podRequests(pod.Spec)
	lcpu, lmem := podLimits(pod.Spec)
	r.cpu, r.lcpu, r.mem, r.lmem = rcpu.MilliValue(), lcpu.MilliValue(), rmem.Value(), lmem.Value()
	if mx != nil {
		ccpu, cmem := currentRes(mx)
		c.cpu, c.mem = ccpu.MilliValue(), cmem.Value()
	}

	return
}

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

func podLimits(spec v1.PodSpec) (resource.Quantity, resource.Quantity) {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	for _, co := range spec.Containers {
		limits := co.Resources.Limits
		if len(limits) == 0 {
			return resource.Quantity{}, resource.Quantity{}
		}
		if limits.Cpu() != nil {
			cpu.Add(*limits.Cpu())
		}
		if limits.Memory() != nil {
			mem.Add(*limits.Memory())
		}
	}
	return *cpu, *mem
}

func podRequests(spec v1.PodSpec) (resource.Quantity, resource.Quantity) {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	for i := range spec.Containers {
		rl := containerRequests(&spec.Containers[i])
		if rl.Cpu() != nil {
			cpu.Add(*rl.Cpu())
		}
		if rl.Memory() != nil {
			mem.Add(*rl.Memory())
		}
	}
	return *cpu, *mem
}

func currentRes(mx *mv1beta1.PodMetrics) (resource.Quantity, resource.Quantity) {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	if mx == nil {
		return *cpu, *mem
	}
	for _, co := range mx.Containers {
		c, m := co.Usage.Cpu(), co.Usage.Memory()
		cpu.Add(*c)
		mem.Add(*m)
	}

	return *cpu, *mem
}

func (*Pod) mapQOS(class v1.PodQOSClass) string {
	// nolint:exhaustive
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
		if po.DeletionTimestamp != nil && po.Status.Reason == NodeUnreachablePodReason {
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

// PosStatus computes pod status.
func PodStatus(pod *v1.Pod) string {
	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodScheduled && condition.Reason == v1.PodReasonSchedulingGated {
			reason = v1.PodReasonSchedulingGated
		}
	}

	var initializing bool
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		var hasRunning bool
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
			}
		}

		if reason == PhaseCompleted && hasRunning {
			if hasPodReadyCondition(pod.Status.Conditions) {
				reason = PhaseRunning
			} else {
				reason = PhaseNotReady
			}
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == NodeUnreachablePodReason {
		reason = PhaseUnknown
	} else if pod.DeletionTimestamp != nil {
		reason = PhaseTerminating
	}

	return reason
}

func hasPodReadyCondition(conditions []v1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}

	return false
}
