package watch

import (
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	v1 "k8s.io/api/core/v1"
	wv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	// PodIndex marker for stored pods.
	PodIndex string = "po"
	podCols         = 11
)

// Connection represents an client api server connection.
type Connection k8s.Connection

// Pod tracks pod activities.
type Pod struct {
	cache.SharedIndexInformer
}

// NewPod returns a new pod.
func NewPod(c Connection, ns string) *Pod {
	return &Pod{
		SharedIndexInformer: wv1.NewPodInformer(c.DialOrDie(), ns, 0, cache.Indexers{}),
	}
}

// List all pods from store in the given namespace.
func (p *Pod) List(ns string) k8s.Collection {
	var res k8s.Collection
	for _, o := range p.GetStore().List() {
		pod := o.(*v1.Pod)
		if ns == "" || pod.Namespace == ns {
			res = append(res, pod)
		}
	}
	return res
}

// Get retrieves a given pod from store.
func (p *Pod) Get(fqn string) (interface{}, error) {
	o, ok, err := p.GetStore().GetByKey(fqn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Pod %s not found", fqn)
	}

	return o, nil
}
