package watch

import (
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ContainerIndex marker for stored containers.
	ContainerIndex string = "co"
	containerCols         = 12
)

// Container tracks container activities.
type Container struct {
	StoreInformer
}

// NewContainer returns a new container.
func NewContainer(po StoreInformer) *Container {
	return &Container{StoreInformer: po}
}

// StartWatching registers container event listener.
func (c *Container) StartWatching(stopCh <-chan struct{}) {}

// Run starts out the informer loop.
func (c *Container) Run(closeCh <-chan struct{}) {}

// Get retrieves a given container from store.
func (c *Container) Get(fqn string, opts metav1.GetOptions) (interface{}, error) {
	o, ok, err := c.GetStore().GetByKey(fqn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Pod<containers> %s not found", fqn)
	}
	po := o.(*v1.Pod)
	cc := make(k8s.Collection, len(po.Spec.InitContainers)+len(po.Spec.Containers))
	toContainers(po, cc)

	return cc, nil
}

// List retrieves alist of containers for a given po from store.
func (c *Container) List(fqn string, opts metav1.ListOptions) k8s.Collection {
	o, ok, err := c.GetStore().GetByKey(fqn)
	if err != nil {
		log.Error().Err(err).Msg("Pod<container>")
		return nil
	}
	if !ok {
		log.Error().Err(fmt.Errorf("Pod<containers> %s not found", fqn)).Msg("Pod<container>")
		return nil
	}
	po := o.(*v1.Pod)
	cc := make(k8s.Collection, len(po.Spec.InitContainers)+len(po.Spec.Containers))
	toContainers(po, cc)

	return cc
}

// ----------------------------------------------------------------------------
// Helpers...

func toContainers(po *v1.Pod, c k8s.Collection) {
	var index int
	for _, co := range po.Spec.InitContainers {
		c[index] = co
		index++
	}
	for _, co := range po.Spec.Containers {
		c[index] = co
		index++
	}
}
