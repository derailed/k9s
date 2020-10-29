package render

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
		HeaderColumn{Name: "%CPU", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "%MEM", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "CPU", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "MEM", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "CPU/R", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "CPU/L", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "MEM/R", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "MEM/L", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "CPU/A", Align: tview.AlignRight, MX: true},
		HeaderColumn{Name: "MEM/A", Align: tview.AlignRight, MX: true},
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

	c, p, a := gatherNodeMX(&no, oo.MX)
	trc, trm, tlc, tlm := new(resource.Quantity), new(resource.Quantity), new(resource.Quantity), new(resource.Quantity)
	for _, p := range oo.Pods {
		rList := podRequests(p.Spec)
		if rList.Cpu() != nil {
			trc.Add(*rList.Cpu())
		}
		if rList.Memory() != nil {
			trm.Add(*rList.Memory())
		}
		lList := podLimits(p.Spec)
		if lList.Cpu() != nil {
			tlc.Add(*lList.Cpu())
		}
		if lList.Memory() != nil {
			tlm.Add(*lList.Memory())
		}
	}
	res := newResources(newResourceList(trc, trm), newResourceList(tlc, tlm))
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
		strconv.Itoa(len(oo.Pods)),
		strconv.Itoa(p.rCPU()),
		strconv.Itoa(p.rMEM()),
		toMc(c.rCPU().MilliValue()),
		toMi(c.rMEM().Value()),
		toMcPerc(res.rCPU(), a.rCPU()),
		toMcPerc(res.lCPU(), a.rCPU()),
		toMiPerc(res.rMEM(), a.rMEM()),
		toMiPerc(res.lMEM(), a.rMEM()),
		toMc(a.rCPU().MilliValue()),
		toMi(a.rMEM().Value()),
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
	Raw  *unstructured.Unstructured
	MX   *mv1beta1.NodeMetrics
	Pods []*v1.Pod
}

// GetObjectKind returns a schema object.
func (n *NodeWithMetrics) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (n *NodeWithMetrics) DeepCopyObject() runtime.Object {
	return n
}

func gatherNodeMX(no *v1.Node, mx *mv1beta1.NodeMetrics) (resources, percentages, resources) {
	c, p, a := newResources(nil, nil), newPercentages(), newResources(no.Status.Allocatable, nil)
	if mx == nil {
		return c, p, a
	}

	c[requestCPU], c[requestMEM] = mx.Usage.Cpu(), mx.Usage.Memory()
	if a.rCPU() != nil {
		p[requestCPU] = percentMc(c.rCPU(), a.rCPU())
	}
	if a.rMEM() != nil {
		p[requestMEM] = percentMi(c.rMEM(), a.rMEM())
	}

	return c, p, a
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
