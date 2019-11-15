package watch

import (
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	wv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	// NodeIndex marker for stored nodes.
	NodeIndex string = "nodes"
	nodeCols         = 12
)

// Node tracks node activities.
type Node struct {
	cache.SharedIndexInformer
}

// NewNode returns a new node.
func NewNode(c k8s.Connection) *Node {
	return &Node{
		SharedIndexInformer: wv1.NewNodeInformer(c.DialOrDie(), 0, cache.Indexers{}),
	}
}

// List all nodes.
func (n *Node) List(_ string, opts metav1.ListOptions) k8s.Collection {
	var res k8s.Collection
	for _, o := range n.GetStore().List() {
		res = append(res, o)
	}

	return res
}

// Get retrieves a given node from store.
func (n *Node) Get(fqn string, opts metav1.GetOptions) (interface{}, error) {
	o, ok, err := n.GetStore().GetByKey(fqn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Node %s not found", fqn)
	}

	return o, nil
}
