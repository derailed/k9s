package resource

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	labelNodeRolePrefix = "node-role.kubernetes.io/"
	nodeLabelRole       = "kubernetes.io/role"
)

// Node tracks a kubernetes resource.
type Node struct {
	*Base
	instance *v1.Node
	metrics  *mv1beta1.NodeMetrics
}

// NewNodeList returns a new resource list.
func NewNodeList(c Connection, _ string) List {
	return NewList(
		NotNamespaced,
		"nodes",
		NewNode(c),
		ViewAccess|DescribeAccess,
	)
}

// NewNode instantiates a new Node.
func NewNode(c Connection) *Node {
	n := &Node{
		Base: &Base{
			Connection: c,
			Resource:   k8s.NewNode(c),
		},
	}
	n.Factory = n

	return n
}

// New builds a new Node instance from a k8s resource.
func (r *Node) New(i interface{}) (Columnar, error) {
	c := NewNode(r.Connection)
	switch instance := i.(type) {
	case *v1.Node:
		c.instance = instance
	case v1.Node:
		c.instance = &instance
	default:
		return nil, fmt.Errorf("Expecting Node but got %T", instance)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c, nil
}

// SetNodeMetrics set the current k8s resource metrics on a given node.
func (r *Node) SetNodeMetrics(m *mv1beta1.NodeMetrics) {
	r.metrics = m
}

// BOZO!!
// // List all resources for a given namespace.
// func (r *Node) List(ns string, opts metav1.ListOptions) (Columnars, error) {
// 	nn, err := r.Resource.List(ns, opts)
// 	if err != nil {
// 		return nil, err
// 	}

// 	cc := make(Columnars, 0, len(nn))
// 	for i := range nn {
// 		node, ok := nn[i].(v1.Node)
// 		if !ok {
// 			return nil, errors.New("Expecting a node resource")
// 		}
// 		no, err := r.New(&node)
// 		if err != nil {
// 			return nil, err
// 		}
// 		cc = append(cc, no)
// 	}

// 	return cc, nil
// }

// Marshal a resource to yaml.
func (r *Node) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		log.Error().Err(err)
		return "", err
	}

	no, ok := i.(*v1.Node)
	if !ok {
		return "", errors.New("Expecting a node resource")
	}
	no.TypeMeta.APIVersion = "v1"
	no.TypeMeta.Kind = "Node"

	return r.marshalObject(no)
}

// Header returns resource header.
func (*Node) Header(ns string) Row {
	return Row{
		"NAME",
		"STATUS",
		"ROLE",
		"VERSION",
		"KERNEL",
		"INTERNAL-IP",
		"EXTERNAL-IP",
		"CPU",
		"MEM",
		"%CPU",
		"%MEM",
		"ACPU",
		"AMEM",
		"AGE",
	}
}

// NumCols designates if column is numerical.
func (*Node) NumCols(n string) map[string]bool {
	return map[string]bool{
		"CPU":  true,
		"MEM":  true,
		"%CPU": true,
		"%MEM": true,
		"ACPU": true,
		"AMEM": true,
	}
}

// Fields returns displayable fields.
func (r *Node) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	no := r.instance
	iIP, eIP := r.getIPs(no.Status.Addresses)
	iIP, eIP = missing(iIP), missing(eIP)

	c, a, p := gatherNodeMX(no, r.metrics)

	sta := make([]string, 10)
	r.status(no.Status, no.Spec.Unschedulable, sta)
	ro := sets.NewString()
	r.findNodeRoles(no, &ro)

	return append(ff,
		no.Name,
		join(sta),
		join(ro.List()),
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
}

// ----------------------------------------------------------------------------
// Helpers...

type metric struct {
	cpu, mem string
}

func noMetric() metric {
	return metric{cpu: NAValue, mem: NAValue}
}

func gatherNodeMX(no *v1.Node, mx *mv1beta1.NodeMetrics) (c metric, a metric, p metric) {
	c, a, p = noMetric(), noMetric(), noMetric()
	if mx == nil {
		return
	}

	cpu := mx.Usage.Cpu().MilliValue()
	mem := k8s.ToMB(mx.Usage.Memory().Value())
	c = metric{
		cpu: ToMillicore(cpu),
		mem: ToMi(mem),
	}

	acpu := no.Status.Allocatable.Cpu().MilliValue()
	amem := k8s.ToMB(no.Status.Allocatable.Memory().Value())
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

func (_ *Node) findNodeRoles(no *v1.Node, roles *sets.String) {
	for k, v := range no.Labels {
		switch {
		case strings.HasPrefix(k, labelNodeRolePrefix):
			if role := strings.TrimPrefix(k, labelNodeRolePrefix); len(role) > 0 {
				roles.Insert(role)
			}
		case k == nodeLabelRole && v != "":
			roles.Insert(v)
		}
	}

	if roles.Len() == 0 {
		roles.Insert(MissingValue)
	}
}

func (*Node) getIPs(addrs []v1.NodeAddress) (iIP, eIP string) {
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

func (*Node) status(status v1.NodeStatus, exempt bool, res []string) {
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
