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
	instance      *v1.Node
	MetricsServer MetricsServer
	metrics       k8s.NodeMetrics
}

// NewNodeList returns a new resource list.
func NewNodeList(c Connection, mx MetricsServer, ns string) List {
	return NewList(
		NotNamespaced,
		"no",
		NewNode(c, mx),
		ViewAccess|DescribeAccess,
	)
}

// NewNode instantiates a new Node.
func NewNode(c Connection, mx MetricsServer) *Node {
	n := &Node{&Base{Connection: c, Resource: k8s.NewNode(c)}, nil, mx, k8s.NodeMetrics{}}
	n.Factory = n

	return n
}

// New builds a new Node instance from a k8s resource.
func (r *Node) New(i interface{}) Columnar {
	c := NewNode(r.Connection, r.MetricsServer)
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

// List all resources for a given namespace.
func (r *Node) List(ns string) (Columnars, error) {
	nn, err := r.Resource.List(ns)
	if err != nil {
		return nil, err
	}

	nodes := make([]v1.Node, 0, len(nn))
	for _, n := range nn {
		nodes = append(nodes, n.(v1.Node))
	}

	mx := make(k8s.NodesMetrics, len(nodes))
	if r.MetricsServer.HasMetrics() {
		nmx, _ := r.MetricsServer.FetchNodesMetrics()
		r.MetricsServer.NodesMetrics(nodes, nmx, mx)
	}

	cc := make(Columnars, 0, len(nodes))
	for i := range nodes {
		no := r.New(&nodes[i]).(*Node)
		no.metrics = mx[nodes[i].Name]
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
		"ROLES",
		"VERSION",
		"KERNEL",
		"INTERNAL-IP",
		"EXTERNAL-IP",
		"CPU",
		"MEM",
		"AVA CPU",
		"AVA MEM",
		"CAP CPU",
		"CAP MEM",
		"AGE",
	}
}

// Fields returns displayable fields.
func (r *Node) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance

	iIP, eIP := r.getIPs(i.Status.Addresses)
	iIP, eIP = missing(iIP), missing(eIP)

	return append(ff,
		i.Name,
		r.status(i),
		missing(strings.Join(findNodeRoles(i), ",")),
		i.Status.NodeInfo.KubeletVersion,
		i.Status.NodeInfo.KernelVersion,
		iIP,
		eIP,
		ToMillicore(r.metrics.CurrentCPU),
		ToMi(r.metrics.CurrentMEM),
		ToMillicore(r.metrics.AvailCPU),
		ToMi(r.metrics.AvailMEM),
		ToMillicore(r.metrics.TotalCPU),
		ToMi(r.metrics.TotalMEM),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
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
