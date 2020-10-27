package render

import (
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	labelNodeRolePrefix = "node-role.kubernetes.io/"
	nodeLabelRole       = "kubernetes.io/role"
)

// Node renders a K8s Node to screen.
type Node struct{}

// ColorerFunc colors a resource row.
func (n Node) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Node) Header(_ string) Header {
	return Header{
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "STATUS"},
		HeaderColumn{Name: "ROLE"},
		HeaderColumn{Name: "VERSION"},
		HeaderColumn{Name: "KERNEL", Wide: true},
		HeaderColumn{Name: "INTERNAL-IP", Wide: true},
		HeaderColumn{Name: "EXTERNAL-IP", Wide: true},
		HeaderColumn{Name: "PODS", Align: tview.AlignRight},
		HeaderColumn{Name: "CPU", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "MEM", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "%CPU", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "%MEM", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "ACPU", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "AMEM", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true, Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (n Node) Render(o interface{}, ns string, r *Row) error {
	oo, ok := o.(*NodeWithMetrics)
	if !ok {
		return fmt.Errorf("Expected *NodeAndMetrics, but got %T", o)
	}

	meta, ok := oo.Raw.Object["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Unable to extract meta")
	}
	na := extractMetaField(meta, "name")
	var no v1.Node
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(oo.Raw.Object, &no)
	if err != nil {
		return err
	}

	iIP, eIP := getIPs(no.Status.Addresses)
	iIP, eIP = missing(iIP), missing(eIP)

	c, a, p := gatherNodeMX(&no, oo.MX)

	statuses := make(sort.StringSlice, 10)
	status(no.Status, no.Spec.Unschedulable, statuses)
	sort.Sort(statuses)
	roles := make(sort.StringSlice, 10)
	nodeRoles(&no, roles)
	sort.Sort(roles)

	r.ID = client.FQN("", na)
	r.Fields = Fields{
		no.Name,
		join(statuses, ","),
		join(roles, ","),
		no.Status.NodeInfo.KubeletVersion,
		no.Status.NodeInfo.KernelVersion,
		iIP,
		eIP,
		strconv.Itoa(oo.PodCount),
		c.cpu,
		c.mem,
		p.cpu,
		p.mem,
		a.cpu,
		a.mem,
		mapToStr(no.Labels),
		asStatus(n.diagnose(statuses)),
		toAge(no.ObjectMeta.CreationTimestamp),
	}

	return nil
}

func (Node) diagnose(ss []string) error {
	if len(ss) == 0 {
		return nil
	}

	var ready bool
	for _, s := range ss {
		if s == "" {
			continue
		}
		if s == "SchedulingDisabled" {
			return errors.New("node is cordoned")
		}
		if s == "Ready" {
			ready = true
		}
	}

	if !ready {
		return errors.New("node is not ready")
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// NodeWithMetrics represents a node with its associated metrics.
type NodeWithMetrics struct {
	Raw      *unstructured.Unstructured
	MX       *mv1beta1.NodeMetrics
	PodCount int
}

// Object returns a k8s Object
func (n *NodeWithMetrics) Object() metav1.Object {
	return n.Raw
}

// GetObjectKind returns a schema object.
func (n *NodeWithMetrics) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (n *NodeWithMetrics) DeepCopyObject() runtime.Object {
	return n
}

func gatherNodeMX(no *v1.Node, mx *mv1beta1.NodeMetrics) (c metric, a metric, p metric) {
	c, a, p = noMetric(), noMetric(), noMetric()
	if mx == nil {
		return
	}

	cpu, mem := mx.Usage.Cpu().MilliValue(), client.ToMB(mx.Usage.Memory().Value())
	c = metric{
		cpu: ToMillicore(cpu),
		mem: ToMi(mem),
	}

	acpu, amem := no.Status.Allocatable.Cpu().MilliValue(), client.ToMB(no.Status.Allocatable.Memory().Value())
	a = metric{
		cpu: ToMillicore(acpu),
		mem: ToMi(amem),
	}

	p = metric{
		cpu: IntToStr(client.ToPercentage(cpu, acpu)),
		mem: IntToStr(client.ToPercentage(mem, amem)),
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
		if index >= len(res) {
			break
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
