package resource

import (
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
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
		"no",
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
func (r *Node) New(i interface{}) Columnar {
	c := NewNode(r.Connection)
	switch instance := i.(type) {
	case *v1.Node:
		c.instance = instance
	case v1.Node:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown Node type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// SetNodeMetrics set the current k8s resource metrics on a given node.
func (r *Node) SetNodeMetrics(m *mv1beta1.NodeMetrics) {
	r.metrics = m
}

// List all resources for a given namespace.
func (r *Node) List(ns string) (Columnars, error) {
	nn, err := r.Resource.List(ns)
	if err != nil {
		return nil, err
	}

	cc := make(Columnars, 0, len(nn))
	for i := range nn {
		node := nn[i].(v1.Node)
		no := r.New(&node).(*Node)
		cc = append(cc, no)
	}

	return cc, nil
}

// Marshal a resource to yaml.
func (r *Node) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		log.Error().Err(err)
		return "", err
	}

	no := i.(*v1.Node)
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
		"ACPU",
		"AMEM",
		"AGE",
	}
}

// Fields returns displayable fields.
func (r *Node) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	no := r.instance
	iIP, eIP := r.getIPs(no.Status.Addresses)
	iIP, eIP = missing(iIP), missing(eIP)

	ccpu, cmem, scpu, smem := NAValue, NAValue, NAValue, NAValue
	if r.metrics != nil {
		var (
			cpu int64
			mem float64
		)

		cpu = r.metrics.Usage.Cpu().MilliValue()
		mem = k8s.ToMB(r.metrics.Usage.Memory().Value())

		acpu := no.Status.Allocatable.Cpu().MilliValue()
		amem := k8s.ToMB(no.Status.Allocatable.Memory().Value())
		ccpu = withPerc(ToMillicore(cpu), AsPerc(toPerc(float64(cpu), float64(acpu))))
		cmem = withPerc(ToMi(mem), AsPerc(toPerc(mem, amem)))
		scpu = ToMillicore(cpu)
		smem = ToMi(mem)
	}

	sta := make([]string, 10)
	r.status(no.Status, no.Spec.Unschedulable, sta)
	ro := make([]string, 10)
	r.nodeRoles(no, ro)

	return append(ff,
		no.Name,
		join(sta, ","),
		join(ro, ","),
		no.Status.NodeInfo.KubeletVersion,
		no.Status.NodeInfo.KernelVersion,
		iIP,
		eIP,
		ccpu,
		cmem,
		scpu,
		smem,
		toAge(no.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func withPerc(v, p string) string {
	return v + " (" + p + ")"
}

func (*Node) nodeRoles(node *v1.Node, res []string) {
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
		index++
	}
}

func findNodeRoles(no *v1.Node) []string {
	roles := sets.NewString()
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

	return roles.List()
}

func (r *Node) podsResources(name string) (v1.ResourceList, v1.ResourceList, error) {
	reqs, limits := v1.ResourceList{}, v1.ResourceList{}
	pods, err := r.Connection.NodePods(name)
	if err != nil {
		return reqs, limits, err
	}
	for _, p := range pods.Items {
		preq, plim := podResources(&p)
		for k, v := range preq {
			if value, ok := reqs[k]; !ok {
				reqs[k] = *v.Copy()
			} else {
				value.Add(v)
				reqs[k] = value
			}
		}
		for k, v := range plim {
			if value, ok := limits[k]; !ok {
				limits[k] = *v.Copy()
			} else {
				value.Add(v)
				limits[k] = value
			}
		}
	}

	return reqs, limits, nil
}

func podResources(pod *v1.Pod) (v1.ResourceList, v1.ResourceList) {
	reqs, limits := v1.ResourceList{}, v1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		addResources(reqs, container.Resources.Requests)
		addResources(limits, container.Resources.Limits)
	}
	// init containers define the minimum of any resource
	for _, container := range pod.Spec.InitContainers {
		maxResources(reqs, container.Resources.Requests)
		maxResources(limits, container.Resources.Limits)
	}

	return reqs, limits
}

// AddResources adds the resources from l2 to l1.
func addResources(l1, l2 v1.ResourceList) {
	for name, quantity := range l2 {
		if value, ok := l1[name]; ok {
			value.Add(quantity)
			l1[name] = value
		} else {
			l1[name] = *quantity.Copy()
		}
	}
}

// MaxResourceList sets list to the greater of l1/l2 for every resource.
func maxResources(l1, l2 v1.ResourceList) {
	for name, quantity := range l2 {
		if value, ok := l1[name]; ok {
			if quantity.Cmp(value) > 0 {
				l1[name] = *quantity.Copy()
			}
		} else {
			l1[name] = *quantity.Copy()
		}
	}
}
