package watch

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	wv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	// NodeIndex marker for stored nodes.
	NodeIndex string = "no"
	nodeCols         = 12
)

// Node tracks node activities.
type Node struct {
	cache.SharedIndexInformer

	client   k8s.Connection
	data     RowEvents
	mxData   k8s.NodesMetrics
	listener TableListenerFn
}

// NewNode returns a new node.
func NewNode(client k8s.Connection) *Node {
	no := Node{
		client: client,
		data:   RowEvents{},
		mxData: k8s.NodesMetrics{},
	}

	if client == nil {
		return &no
	}

	no.SharedIndexInformer = wv1.NewNodeInformer(
		client.DialOrDie(),
		0,
		cache.Indexers{},
	)
	no.SharedIndexInformer.AddEventHandler(&no)

	return &no
}

// List all nodes.
func (n *Node) List(_ string) (k8s.Collection, error) {
	var res k8s.Collection
	for _, o := range n.GetStore().List() {
		res = append(res, o)
	}
	return res, nil
}

// Get retrieves a given node from store.
func (n *Node) Get(fqn string) (interface{}, error) {
	o, ok, err := n.GetStore().GetByKey(fqn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Node %s not found", fqn)
	}

	return o, nil
}

// SetListener registers event recipient.
func (n *Node) SetListener(_ string, cb TableListenerFn) {
	n.listener = cb
	n.fireChanged()
}

// UnsetListener unregister event recipient.
func (n *Node) UnsetListener(_ string) {
	n.listener = nil
}

func (n *Node) fetchMetrics() {
	if n.client == nil {
		return
	}

	client := k8s.NewMetricsServer(n.client)
	mx, err := client.FetchNodesMetrics()
	if err != nil {
		log.Error().Err(err).Msg("Node metrics failed")
		return
	}
	client.NodesMetrics(n.GetStore().List(), mx, n.mxData)
}

// Data return current data.
func (n *Node) tableData() TableData {
	return TableData{
		Header:    n.header(),
		Rows:      n.data,
		Namespace: NotNamespaced,
	}
}

func (n *Node) fireChanged() {
	if cb := n.listener; cb != nil {
		cb(n.tableData())
	}
}

// OnAdd notify node added.
func (n *Node) OnAdd(obj interface{}) {
	no := obj.(*v1.Node)
	n.fetchMetrics()
	ff := make(Row, nodeCols)
	n.fields(no, ff)
	fqn := MetaFQN(no.ObjectMeta)
	log.Debug().Msgf("Node Added %s", fqn)

	n.data[fqn] = &RowEvent{
		Action: watch.Added,
		Fields: ff,
		Deltas: make(Row, len(ff)),
	}

	if n.HasSynced() {
		n.fireChanged()
	}
}

// OnUpdate notify node updated.
func (n *Node) OnUpdate(oldObj, newObj interface{}) {
	ono, nno := oldObj.(*v1.Node), newObj.(*v1.Node)
	k1 := MetaFQN(ono.ObjectMeta)
	k2 := MetaFQN(nno.ObjectMeta)

	ff := make(Row, nodeCols)
	n.fields(nno, ff)
	if re, ok := n.data[k1]; ok {
		re.Action = watch.Modified
		re.Deltas = re.Fields
		re.Fields = ff
	}
	n.data[k2] = &RowEvent{
		Action: watch.Added,
		Fields: ff,
		Deltas: make(Row, len(ff)),
	}
	n.fireChanged()
}

// OnDelete notify node was deleted.
func (n *Node) OnDelete(obj interface{}) {
	po := obj.(*v1.Node)
	key := MetaFQN(po.ObjectMeta)

	log.Debug().Msgf("Node Delete %s", key)

	delete(n.data, key)
	n.fireChanged()
}

// Header returns resource header.
func (*Node) header() Row {
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
func (n *Node) fields(no *v1.Node, r Row) {
	col := 0
	r[col] = no.Name
	col++
	sta := make([]string, 10)
	n.status(no.Status, no.Spec.Unschedulable, sta)
	r[col] = join(sta, ",")
	col++
	ro := make([]string, 10)
	n.nodeRoles(no, ro)
	r[col] = join(ro, ",")
	col++
	r[col] = no.Status.NodeInfo.KubeletVersion
	col++
	r[col] = no.Status.NodeInfo.KernelVersion
	col++
	iIP, eIP := n.getIPs(no.Status.Addresses)
	iIP, eIP = missing(iIP), missing(eIP)
	r[col], r[col+1] = iIP, eIP
	col += 2

	fqn := MetaFQN(no.ObjectMeta)
	n.fetchMetrics()
	mx := n.mxData[fqn]
	r[col] = withPerc(
		ToMillicore(mx.CurrentCPU),
		AsPerc(toPerc(float64(mx.CurrentCPU), float64(mx.AvailCPU))),
	)
	col++
	r[col] = withPerc(
		ToMi(mx.CurrentMEM),
		AsPerc(toPerc(mx.CurrentMEM, mx.AvailMEM)),
	)
	col++
	r[col] = ToMillicore(mx.AvailCPU)
	col++
	r[col] = ToMi(mx.AvailMEM)
	col++
	r[col] = toAge(no.ObjectMeta.CreationTimestamp)
}

// ----------------------------------------------------------------------------
// Helpers...

const (
	labelNodeRolePrefix = "node-role.kubernetes.io/"
	nodeLabelRole       = "kubernetes.io/role"
)

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
		index++
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
