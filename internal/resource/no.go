package resource

import (
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
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
		"ROLE",
		"VERSION",
		"KERNEL",
		"INTERNAL-IP",
		"EXTERNAL-IP",
		"CPU",
		"MEM",
		"RCPU",
		"RMEM",
		"ACPU",
		"AMEM",
		"AGE",
	}
}

// Fields returns displayable fields.
func (r *Node) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	i := r.instance
	iIP, eIP := r.getIPs(i.Status.Addresses)
	iIP, eIP = missing(iIP), missing(eIP)

	reqs, _, err := r.fetchReqLimit(i.Name)
	if err != nil {
		if !errors.IsForbidden(err) {
			log.Warn().Msgf("User is not authorized to list pods on nodes: %v", err)
		}
		log.Error().Msgf("%#v", err)
	}

	rcpu, rmem := reqs["cpu"], reqs["memory"]

	pcpur := toPerc(float64(rcpu.MilliValue()), float64(r.metrics.AvailCPU))
	pmemr := toPerc(k8s.ToMB(rmem.Value()), float64(r.metrics.AvailMEM))

	return append(ff,
		i.Name,
		r.status(i),
		r.nodeRoles(i),
		i.Status.NodeInfo.KubeletVersion,
		i.Status.NodeInfo.KernelVersion,
		iIP,
		eIP,
		withPerc(ToMillicore(r.metrics.CurrentCPU), asPerc(toPerc(float64(r.metrics.CurrentCPU), float64(r.metrics.AvailCPU)))),
		withPerc(ToMi(r.metrics.CurrentMEM), asPerc(toPerc(r.metrics.CurrentMEM, r.metrics.AvailMEM))),
		withPerc(rcpu.String(), asPerc(pcpur)),
		withPerc(rmem.String(), asPerc(pmemr)),
		ToMillicore(r.metrics.AvailCPU),
		ToMi(r.metrics.AvailMEM),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

func withPerc(v, p string) string {
	return v + " (" + p + ")"
}

// ----------------------------------------------------------------------------
// Helpers...

func (*Node) nodeRoles(node *v1.Node) string {
	const (
		labelNodeRolePrefix = "node-role.kubernetes.io/"
		nodeLabelRole       = "kubernetes.io/role"
	)

	roles := sets.NewString()
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, labelNodeRolePrefix):
			if role := strings.TrimPrefix(k, labelNodeRolePrefix); len(role) > 0 {
				roles.Insert(role)
			}

		case k == nodeLabelRole && v != "":
			roles.Insert(v)
		}
	}

	if len(roles) == 0 {
		return MissingValue
	}
	return strings.Join(roles.List(), ",")
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

func (r *Node) fetchReqLimit(name string) (req, lim v1.ResourceList, err error) {
	reqs, limits := map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}

	pods, err := r.Connection.ValidPods(name)
	for _, p := range pods {
		preq, plim := podRequestsAndLimits(&p)
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

func podRequestsAndLimits(pod *v1.Pod) (reqs, limits v1.ResourceList) {
	reqs, limits = map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}

	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
		addResourceList(limits, container.Resources.Limits)
	}
	// init containers define the minimum of any resource
	for _, container := range pod.Spec.InitContainers {
		maxResourceList(reqs, container.Resources.Requests)
		maxResourceList(limits, container.Resources.Limits)
	}
	return
}

// addResourceList adds the resources in newList to list
func addResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = *quantity.Copy()
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}

// maxResourceList sets list to the greater of list/newList for every resource
// either list
func maxResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = *quantity.Copy()
			continue
		} else {
			if quantity.Cmp(value) > 0 {
				list[name] = *quantity.Copy()
			}
		}
	}
}
