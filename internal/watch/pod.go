package watch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	wv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// PodIndex marker for stored pods.
const PodIndex = "pods"

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
func (p *Pod) List(ns string, opts metav1.ListOptions) k8s.Collection {
	var res k8s.Collection
	var nodeSelector bool
	if strings.Contains(opts.FieldSelector, "spec.nodeName") {
		nodeSelector = true
	}
	for _, o := range p.GetStore().List() {
		pod, ok := o.(*v1.Pod)
		if !ok {
			log.Error().Err(errors.New("Expecting a pod"))
			return res
		}
		if ns != "" && pod.Namespace != ns {
			continue
		}
		if nodeSelector {
			if !matchesNode(pod.Spec.NodeName, toSelector(opts.FieldSelector)) {
				continue
			}
		} else if !matchesLabels(pod.ObjectMeta.Labels, toSelector(opts.LabelSelector)) {
			continue
		}
		res = append(res, pod)
	}
	return res
}

// Get retrieves a given pod from store.
func (p *Pod) Get(fqn string, opts metav1.GetOptions) (interface{}, error) {
	o, ok, err := p.GetStore().GetByKey(fqn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Pod %s not found", fqn)
	}

	return o, nil
}
