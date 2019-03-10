package resource

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

const (
	labelNodeRolePrefix = "node-role.kubernetes.io/"
	nodeLabelRole       = "kubernetes.io/role"
)

// Node tracks a kubernetes resource.
type Node struct {
	*Base
	instance  *v1.Node
	metricSvc MetricsIfc
	metrics   k8s.Metric
}

// NewNodeList returns a new resource list.
func NewNodeList(ns string) List {
	return NewNodeListWithArgs(ns, NewNode())
}

// NewNodeListWithArgs returns a new resource list.
func NewNodeListWithArgs(ns string, res Resource) List {
	return newList(NotNamespaced, "no", res, ViewAccess|DescribeAccess)
}

// NewNode instantiates a new Endpoint.
func NewNode() *Node {
	return NewNodeWithArgs(k8s.NewNode(), k8s.NewMetricsServer())
}

// NewNodeWithArgs instantiates a new Endpoint.
func NewNodeWithArgs(r k8s.Res, mx MetricsIfc) *Node {
	ep := &Node{
		metricSvc: mx,
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*Node) NewInstance(i interface{}) Columnar {
	cm := NewNode()
	switch i.(type) {
	case *v1.Node:
		cm.instance = i.(*v1.Node)
	case v1.Node:
		ii := i.(v1.Node)
		cm.instance = &ii
	default:
		log.Fatal().Msgf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// List all resources for a given namespace.
func (r *Node) List(ns string) (Columnars, error) {
	ii, err := r.caller.List(AllNamespaces)
	if err != nil {
		return nil, err
	}

	nn := make([]v1.Node, len(ii))
	for k, i := range ii {
		nn[k] = i.(v1.Node)
	}

	cc := make(Columnars, 0, len(nn))
	mx, err := r.metricSvc.PerNodeMetrics(nn)
	if err != nil {
		log.Warn().Msgf("No metrics: %#v", err)
	}

	for i := 0; i < len(nn); i++ {
		n := r.NewInstance(&nn[i]).(*Node)
		if err == nil {
			n.metrics = mx[nn[i].Name]
		}
		cc = append(cc, n)
	}
	return cc, nil
}

// Marshal a resource to yaml.
func (r *Node) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
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
		"ROLES",
		"VERSION",
		"INTERNAL-IP",
		"EXTERNAL-IP",
		"CPU",
		"MEM",
		"AVAILABLE_CPU",
		"AVAILABLE_MEM",
		"AGE",
	}
}

// Fields returns displayable fields.
func (r *Node) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance

	status := r.status(i)
	iIP, eIP := r.getIPs(i.Status.Addresses)
	iIP, eIP = missing(iIP), missing(eIP)

	roles := missing(strings.Join(findNodeRoles(i), ","))
	cpu, mem, acpu, amem := na(r.metrics.CPU), na(r.metrics.Mem), na(r.metrics.AvailCPU), na(r.metrics.AvailMem)

	return append(ff,
		i.Name,
		status,
		roles,
		i.Status.NodeInfo.KernelVersion,
		iIP,
		eIP,
		cpu,
		mem,
		acpu,
		amem,
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*Node) ExtFields() Properties {
	return Properties{}
}

// Helpers...

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

func (r *Node) status(i *v1.Node) string {
	conditionMap := make(map[v1.NodeConditionType]*v1.NodeCondition)
	NodeAllConditions := []v1.NodeConditionType{v1.NodeReady}
	for n := range i.Status.Conditions {
		cond := i.Status.Conditions[n]
		conditionMap[cond.Type] = &cond
	}
	var status []string
	for _, validCondition := range NodeAllConditions {
		if condition, ok := conditionMap[validCondition]; ok {
			if condition.Status == v1.ConditionTrue {
				status = append(status, string(condition.Type))
			} else {
				status = append(status, "Not"+string(condition.Type))
			}
		}
	}
	if len(status) == 0 {
		status = append(status, "Unknown")
	}
	if i.Spec.Unschedulable {
		status = append(status, "SchedulingDisabled")
	}
	return strings.Join(status, ",")
}

func findNodeRoles(i *v1.Node) []string {
	roles := sets.NewString()
	for k, v := range i.Labels {
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
