package render

import (
	"fmt"
	"strings"

	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	labelNodeRolePrefix = "node-role.kubernetes.io/"
	nodeLabelRole       = "kubernetes.io/role"
)

// NodeWithMetrics represents a resourve object with usage metrics.
type NodeWithMetrics interface {
	Object() runtime.Object
	Metrics() *mv1beta1.NodeMetrics
	Pods() []*v1.Pod
}

// Node renders a K8s Node to screen.
type Node struct{}

// ColorerFunc colors a resource row.
func (Node) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Node) Header(_ string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "STATUS"},
		Header{Name: "ROLE"},
		Header{Name: "VERSION"},
		Header{Name: "KERNEL"},
		Header{Name: "INTERNAL-IP"},
		Header{Name: "EXTERNAL-IP"},
		Header{Name: "CPU", Align: tview.AlignRight},
		Header{Name: "MEM", Align: tview.AlignRight},
		Header{Name: "%CPU", Align: tview.AlignRight},
		Header{Name: "%MEM", Align: tview.AlignRight},
		Header{Name: "ACPU", Align: tview.AlignRight},
		Header{Name: "AMEM", Align: tview.AlignRight},
		Header{Name: "AGE", Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (n Node) Render(o interface{}, ns string, r *Row) error {
	oo, ok := o.(NodeWithMetrics)
	if !ok {
		return fmt.Errorf("Expected NodeAndMetrics, but got %T", o)
	}

	var no v1.Node
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(oo.Object().(*unstructured.Unstructured).Object, &no)
	if err != nil {
		log.Error().Err(err).Msg("Converting Node")
		return err
	}

	iIP, eIP := getIPs(no.Status.Addresses)
	iIP, eIP = missing(iIP), missing(eIP)

	c, a, p := gatherNodeMX(&no, oo.Metrics())

	sta := make([]string, 10)
	status(no.Status, no.Spec.Unschedulable, sta)
	ro := make([]string, 10)
	nodeRoles(&no, ro)

	r.ID = MetaFQN(no.ObjectMeta)
	r.Fields = make(Fields, 0, len(n.Header(ns)))
	r.Fields = append(r.Fields,
		no.Name,
		join(sta, ","),
		join(ro, ","),
		no.Status.NodeInfo.KubeletVersion,
		no.Status.NodeInfo.KernelVersion,
		iIP,
		eIP,
		c.cpu,
		c.mem,
		p.cpu,
		p.mem,
		a.cpu,
		a.mem,
		toAge(no.ObjectMeta.CreationTimestamp),
	)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func gatherNodeMX(no *v1.Node, mx *mv1beta1.NodeMetrics) (c metric, a metric, p metric) {
	c, a, p = noMetric(), noMetric(), noMetric()
	if mx == nil {
		return
	}

	cpu := mx.Usage.Cpu().MilliValue()
	mem := ToMB(mx.Usage.Memory().Value())
	c = metric{
		cpu: ToMillicore(cpu),
		mem: ToMi(mem),
	}

	acpu := no.Status.Allocatable.Cpu().MilliValue()
	amem := ToMB(no.Status.Allocatable.Memory().Value())
	a = metric{
		cpu: ToMillicore(acpu),
		mem: ToMi(amem),
	}

	p = metric{
		cpu: AsPerc(toPerc(float64(cpu), float64(acpu))),
		mem: AsPerc(toPerc(mem, amem)),
	}

	return
}

func nodeRoles(node *v1.Node, res []string) {
	index := 0
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, labelNodeRolePrefix):
			if role := strings.TrimPrefix(k, labelNodeRolePrefix); len(role) > 0 {
				res[index] = role
				index++
			}
		case k == nodeLabelRole && v != "":
			res[index] = v
			index++
		}
	}

	if empty(res) {
		res[index] = MissingValue
	}
}

func getIPs(addrs []v1.NodeAddress) (iIP, eIP string) {
	for _, a := range addrs {
		switch a.Type {
		case v1.NodeExternalIP:
			eIP = a.Address
		case v1.NodeInternalIP:
			iIP = a.Address
		}
	}

	return
}

func status(status v1.NodeStatus, exempt bool, res []string) {
	var index int
	conditions := make(map[v1.NodeConditionType]*v1.NodeCondition)
	for n := range status.Conditions {
		cond := status.Conditions[n]
		conditions[cond.Type] = &cond
	}

	validConditions := []v1.NodeConditionType{v1.NodeReady}
	for _, validCondition := range validConditions {
		condition, ok := conditions[validCondition]
		if !ok {
			continue
		}
		neg := ""
		if condition.Status != v1.ConditionTrue {
			neg = "Not"
		}
		res[index] = neg + string(condition.Type)
		index++

	}
	if len(res) == 0 {
		res[index] = "Unknown"
		index++
	}
	if exempt {
		res[index] = "SchedulingDisabled"
	}
}

func empty(s []string) bool {
	for _, v := range s {
		if len(v) != 0 {
			return false
		}
	}
	return true
}
