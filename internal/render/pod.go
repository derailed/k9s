// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	// NodeUnreachablePodReason is reason and message set on a pod when its state
	// cannot be confirmed as kubelet is unresponsive on the node it is (was) running.
	NodeUnreachablePodReason = "NodeLost" // k8s.io/kubernetes/pkg/util/node.NodeUnreachablePodReason
	vulIdx                   = 2
)

const (
	PhaseTerminating            = "Terminating"
	PhaseInitialized            = "Initialized"
	PhaseRunning                = "Running"
	PhaseNotReady               = "NoReady"
	PhaseCompleted              = "Completed"
	PhaseContainerCreating      = "ContainerCreating"
	PhasePodInitializing        = "PodInitializing"
	PhaseUnknown                = "Unknown"
	PhaseCrashLoop              = "CrashLoopBackOff"
	PhaseError                  = "Error"
	PhaseImagePullBackOff       = "ImagePullBackOff"
	PhaseOOMKilled              = "OOMKilled"
	PhasePending                = "Pending"
	PhaseContainerStatusUnknown = "ContainerStatusUnknown"
	PhaseEvicted                = "Evicted"
)

var defaultPodHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "VS", Attrs: model1.Attrs{VS: true}},
	model1.HeaderColumn{Name: "PF"},
	model1.HeaderColumn{Name: "READY"},
	model1.HeaderColumn{Name: "STATUS"},
	model1.HeaderColumn{Name: "RESTARTS", Attrs: model1.Attrs{Align: tview.AlignRight}},
	model1.HeaderColumn{Name: "LAST RESTART", Attrs: model1.Attrs{Align: tview.AlignRight, Time: true, Wide: true}},
	model1.HeaderColumn{Name: "CPU", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "MEM", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "CPU/RL", Attrs: model1.Attrs{Align: tview.AlignRight, Wide: true}},
	model1.HeaderColumn{Name: "MEM/RL", Attrs: model1.Attrs{Align: tview.AlignRight, Wide: true}},
	model1.HeaderColumn{Name: "%CPU/R", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "%CPU/L", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "%MEM/R", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "%MEM/L", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "IP"},
	model1.HeaderColumn{Name: "NODE"},
	model1.HeaderColumn{Name: "SERVICE-ACCOUNT", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "NOMINATED NODE", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "READINESS GATES", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "QOS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// Pod renders a K8s Pod to screen.
type Pod struct {
	*Base
}

// NewPod returns a new instance.
func NewPod() *Pod {
	return &Pod{
		Base: new(Base),
	}
}

// ColorerFunc colors a resource row.
func (*Pod) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)

		idx, ok := h.IndexOf("STATUS", true)
		if !ok {
			return c
		}
		status := strings.TrimSpace(re.Row.Fields[idx])
		switch status {
		case Pending, ContainerCreating:
			c = model1.PendingColor
		case PodInitializing:
			c = model1.AddColor
		case Initialized:
			c = model1.HighlightColor
		case Completed:
			c = model1.CompletedColor
		case Running:
			if c != model1.ErrColor {
				c = model1.StdColor
			}
		case Terminating:
			c = model1.KillColor
		}

		return c
	}
}

// Header returns a header row.
func (p *Pod) Header(string) model1.Header {
	return p.doHeader(defaultPodHeader)
}

// Render renders a K8s resource to screen.
func (p *Pod) Render(o any, _ string, row *model1.Row) error {
	pwm, ok := o.(*PodWithMetrics)
	if !ok {
		return fmt.Errorf("expected PodWithMetrics, but got %T", o)
	}
	if err := p.defaultRow(pwm, row); err != nil {
		return err
	}
	if p.specs.isEmpty() {
		return nil
	}
	cols, err := p.specs.realize(pwm.Raw, defaultPodHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (p *Pod) defaultRow(pwm *PodWithMetrics, row *model1.Row) error {
	var st v1.PodStatus
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(pwm.Raw.Object["status"].(map[string]any), &st); err != nil {
		return err
	}
	spec := new(v1.PodSpec)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(pwm.Raw.Object["spec"].(map[string]any), spec); err != nil {
		return err
	}

	dt := pwm.Raw.GetDeletionTimestamp()
	_, _, irc, _ := p.Statuses(st.InitContainerStatuses)
	cr, _, rc, lr := p.Statuses(st.ContainerStatuses)

	var ccmx []mv1beta1.ContainerMetrics
	if pwm.MX != nil {
		ccmx = pwm.MX.Containers
	}
	c, r := gatherCoMX(spec, ccmx)
	phase := p.Phase(dt, spec, &st)

	ns, n := pwm.Raw.GetNamespace(), pwm.Raw.GetName()

	row.ID = client.FQN(ns, n)
	row.Fields = model1.Fields{
		ns,
		n,
		computeVulScore(ns, pwm.Raw.GetLabels(), spec),
		"●",
		strconv.Itoa(cr) + "/" + strconv.Itoa(len(spec.Containers)),
		phase,
		strconv.Itoa(rc + irc),
		ToAge(lr),
		toMc(c.cpu),
		toMi(c.mem),
		toMc(r.cpu) + ":" + toMc(r.lcpu),
		toMi(r.mem) + ":" + toMi(r.lmem),
		client.ToPercentageStr(c.cpu, r.cpu),
		client.ToPercentageStr(c.cpu, r.lcpu),
		client.ToPercentageStr(c.mem, r.mem),
		client.ToPercentageStr(c.mem, r.lmem),
		na(st.PodIP),
		na(spec.NodeName),
		na(spec.ServiceAccountName),
		asNominated(st.NominatedNodeName),
		asReadinessGate(spec, &st),
		p.mapQOS(st.QOSClass),
		mapToStr(pwm.Raw.GetLabels()),
		AsStatus(p.diagnose(phase, cr, len(st.ContainerStatuses))),
		ToAge(pwm.Raw.GetCreationTimestamp()),
	}

	return nil
}

func (*Pod) diagnose(phase string, cr, ct int) error {
	if phase == Completed {
		return nil
	}
	if cr != ct || ct == 0 {
		return fmt.Errorf("container ready check failed: %d of %d", cr, ct)
	}
	if phase == Terminating {
		return fmt.Errorf("pod is terminating")
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

func asReadinessGate(spec *v1.PodSpec, st *v1.PodStatus) string {
	if len(spec.ReadinessGates) == 0 {
		return MissingValue
	}

	var trueConditions int
	for _, readinessGate := range spec.ReadinessGates {
		conditionType := readinessGate.ConditionType
		for _, condition := range st.Conditions {
			if condition.Type == conditionType {
				if condition.Status == "True" {
					trueConditions++
				}
				break
			}
		}
	}

	return strconv.Itoa(trueConditions) + "/" + strconv.Itoa(len(spec.ReadinessGates))
}

// PodWithMetrics represents a pod and its metrics.
type PodWithMetrics struct {
	Raw *unstructured.Unstructured
	MX  *mv1beta1.PodMetrics
}

// GetObjectKind returns a schema object.
func (*PodWithMetrics) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (p *PodWithMetrics) DeepCopyObject() runtime.Object {
	return p
}

func gatherCoMX(spec *v1.PodSpec, ccmx []mv1beta1.ContainerMetrics) (c, r metric) {
	cc := make([]v1.Container, 0, len(spec.InitContainers)+len(spec.Containers))
	cc = append(cc, filterSidecarCO(spec.InitContainers)...)
	cc = append(cc, spec.Containers...)

	rcpu, rmem := cosRequests(cc)
	r.cpu, r.mem = rcpu.MilliValue(), rmem.Value()

	lcpu, lmem := cosLimits(cc)
	r.lcpu, r.lmem = lcpu.MilliValue(), lmem.Value()

	ccpu, cmem := currentRes(ccmx)
	c.cpu, c.mem = ccpu.MilliValue(), cmem.Value()

	return
}

func cosLimits(cc []v1.Container) (cpuQ, memQ resource.Quantity) {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	for i := range cc {
		limits := cc[i].Resources.Limits
		if len(limits) == 0 {
			continue
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

func cosRequests(cc []v1.Container) (cpuQ, memQ resource.Quantity) {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	for i := range cc {
		co := cc[i]
		rl := containerRequests(&co)
		if rl.Cpu() != nil {
			cpu.Add(*rl.Cpu())
		}
		if rl.Memory() != nil {
			mem.Add(*rl.Memory())
		}
	}

	return *cpu, *mem
}

func currentRes(ccmx []mv1beta1.ContainerMetrics) (cpuQ, memQ resource.Quantity) {
	cpu, mem := new(resource.Quantity), new(resource.Quantity)
	if ccmx == nil {
		return *cpu, *mem
	}
	for _, co := range ccmx {
		c, m := co.Usage.Cpu(), co.Usage.Memory()
		cpu.Add(*c)
		mem.Add(*m)
	}

	return *cpu, *mem
}

func (*Pod) mapQOS(class v1.PodQOSClass) string {
	//nolint:exhaustive
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
func (*Pod) Statuses(cc []v1.ContainerStatus) (cr, ct, rc int, latest metav1.Time) {
	for i := range cc {
		if cc[i].State.Terminated != nil {
			ct++
		}
		if cc[i].Ready {
			cr++
		}
		rc += int(cc[i].RestartCount)

		if t := cc[i].LastTerminationState.Terminated; t != nil {
			ts := cc[i].LastTerminationState.Terminated.FinishedAt
			if latest.IsZero() || ts.After(latest.Time) {
				latest = ts
			}
		}
	}

	return
}

// Phase reports the given pod phase.
func (p *Pod) Phase(dt *metav1.Time, spec *v1.PodSpec, st *v1.PodStatus) string {
	status := string(st.Phase)
	if st.Reason != "" {
		if dt != nil && st.Reason == NodeUnreachablePodReason {
			return "Unknown"
		}
		status = st.Reason
	}

	status, ok := p.initContainerPhase(spec, st, status)
	if ok {
		return status
	}

	status, ok = p.containerPhase(st, status)
	if ok && status == Completed {
		status = Running
	}
	if dt == nil {
		return status
	}

	return Terminating
}

func (*Pod) containerPhase(st *v1.PodStatus, status string) (string, bool) {
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

func (*Pod) initContainerPhase(spec *v1.PodSpec, pst *v1.PodStatus, status string) (string, bool) {
	count := len(spec.InitContainers)
	rs := make(map[string]bool, count)
	for i := range spec.InitContainers {
		co := spec.InitContainers[i]
		rs[co.Name] = restartableInitCO(co.RestartPolicy)
	}
	for i := range pst.InitContainerStatuses {
		if s := checkInitContainerStatus(&pst.InitContainerStatuses[i], i, count, rs[pst.InitContainerStatuses[i].Name]); s != "" {
			return s, true
		}
	}

	return status, false
}

// ----------------------------------------------------------------------------
// Helpers..

func checkInitContainerStatus(cs *v1.ContainerStatus, count, initCount int, restartable bool) string {
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
	case restartable && cs.Started != nil && *cs.Started:
		if cs.Ready {
			return ""
		}
	case cs.State.Waiting != nil && cs.State.Waiting.Reason != "" && cs.State.Waiting.Reason != "PodInitializing":
		return "Init:" + cs.State.Waiting.Reason
	}

	return "Init:" + strconv.Itoa(count) + "/" + strconv.Itoa(initCount)
}

// PodStatus computes pod status.
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
			if container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && container.State.Waiting.Reason != "" && container.State.Waiting.Reason != "PodInitializing":
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
			switch {
			case container.State.Waiting != nil && container.State.Waiting.Reason != "":
				reason = container.State.Waiting.Reason
			case container.State.Terminated != nil && container.State.Terminated.Reason != "":
				reason = container.State.Terminated.Reason
			case container.State.Terminated != nil && container.State.Terminated.Reason == "":
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			case container.Ready && container.State.Running != nil:
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

func restartableInitCO(p *v1.ContainerRestartPolicy) bool {
	return p != nil && *p == v1.ContainerRestartPolicyAlways
}

func filterSidecarCO(cc []v1.Container) []v1.Container {
	rcc := make([]v1.Container, 0, len(cc))
	for i := range cc {
		c := cc[i]
		if c.RestartPolicy != nil && *c.RestartPolicy == v1.ContainerRestartPolicyAlways {
			rcc = append(rcc, c)
		}
	}

	return rcc
}
