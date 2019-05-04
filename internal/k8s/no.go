package k8s

import (
	"time"

	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Node represents a Kubernetes node.
type Node struct {
	*base
	Connection
}

// NewNode returns a new Node.
func NewNode(c Connection) *Node {
	return &Node{&base{}, c}
}

// Get a node.
func (n *Node) Get(_, name string) (interface{}, error) {
	return n.DialOrDie().CoreV1().Nodes().Get(name, metav1.GetOptions{})
}

// List all nodes on the cluster.
func (n *Node) List(_ string) (Collection, error) {
	defer func(t time.Time) {
		log.Debug().Msgf("List Node %v", time.Since(t))
	}(time.Now())

	opts := metav1.ListOptions{
		LabelSelector: n.labelSelector,
		FieldSelector: n.fieldSelector,
	}
	rr, err := n.DialOrDie().CoreV1().Nodes().List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a node.
func (n *Node) Delete(_, name string) error {
	return n.DialOrDie().CoreV1().Nodes().Delete(name, nil)
}
