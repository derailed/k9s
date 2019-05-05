package watch

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	// ContainerIndex marker for stored containers.
	ContainerIndex string = "co"
	containerCols         = 12
)

// Container tracks container activities.
type Container struct {
	CallbackInformer

	data      RowEvents
	mxData    k8s.PodMetrics
	listener  TableListenerFn
	activeFQN *string
}

// NewContainer returns a new container.
func NewContainer(po CallbackInformer) *Container {
	co := Container{
		CallbackInformer: po,
		data:             RowEvents{},
	}
	po.AddEventHandler(&co)

	return &co
}

// Run starts out the informer loop.
func (c *Container) Run(closeCh <-chan struct{}) {}

// Get retrieves a given container from store.
func (c *Container) Get(fqn string) (interface{}, error) {
	o, ok, err := c.GetStore().GetByKey(fqn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Container %s not found", fqn)
	}

	return o, nil
}

// List retrieves a given containers from store.
func (c *Container) List(fqn string) (k8s.Collection, error) {
	o, ok, err := c.GetStore().GetByKey(fqn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Pod<containers> %s not found", fqn)
	}

	po := o.(*v1.Pod)
	var cc k8s.Collection
	for i := 0; i < len(po.Spec.InitContainers); i++ {
		cc = append(cc, &po.Spec.InitContainers[i])
	}
	for i := 0; i < len(po.Spec.Containers); i++ {
		cc = append(cc, &po.Spec.Containers[i])
	}

	return cc, nil
}

// SetListener registers event recipient.
func (c *Container) SetListener(fqn string, cb TableListenerFn) {
	c.listener, c.activeFQN = cb, &fqn
	o, err := c.Get(fqn)
	if err != nil {
		log.Error().Err(err).Msgf("Pod `%q not found", fqn)
		return
	}
	// Clear out all rows
	for k := range c.data {
		delete(c.data, k)
	}
	c.updateData(watch.Added, o.(*v1.Pod))
	c.fireChanged()
}

// UnsetListener unregister event recipient.
func (c *Container) UnsetListener(_ string) {
	c.listener, c.activeFQN = nil, nil
}

// Data return current data.
func (c *Container) tableData(ns string) TableData {
	return TableData{
		Header:    c.header(),
		Rows:      c.data,
		Namespace: ns,
	}
}

func (c *Container) fireChanged() {
	if cb := c.listener; cb != nil {
		cb(c.tableData(NotNamespaced))
	}
}

// StartWatching registers container event listener.
func (c *Container) StartWatching(stopCh <-chan struct{}) {}

// OnAdd notify container added.
func (c *Container) OnAdd(obj interface{}) {
	if c.activeFQN == nil {
		return
	}

	po := obj.(*v1.Pod)
	fqn := MetaFQN(po.ObjectMeta)
	if fqn != *c.activeFQN {
		return
	}

	log.Debug().Msgf("Pod Added %s", fqn)
	// ff := make(Row, containerCols)
	c.updateData(watch.Added, po)

	if c.HasSynced() {
		c.fireChanged()
	}
}

// OnUpdate notify container updated.
func (c *Container) OnUpdate(oldObj, newObj interface{}) {
	if c.activeFQN == nil {
		return
	}

	opo, npo := oldObj.(*v1.Pod), newObj.(*v1.Pod)
	k1 := MetaFQN(opo.ObjectMeta)
	k2 := MetaFQN(npo.ObjectMeta)

	if k1 != *c.activeFQN && k2 != *c.activeFQN {
		return
	}

	log.Debug().Msgf("Pod Updated %#v - %#v", opo.Name, npo.Name)

	// Check if this is a rollout
	if k1 != k2 {
		c.updateData(watch.Modified, opo)
	} else {
		c.updateData(watch.Modified, npo)
	}
	c.fireChanged()
}

// OnDelete notify container was deleted.
func (c *Container) OnDelete(obj interface{}) {
	if c.activeFQN == nil {
		return
	}

	po := obj.(*v1.Pod)
	fqn := MetaFQN(po.ObjectMeta)
	if fqn != *c.activeFQN {
		return
	}
	log.Debug().Msgf("Pod Deleted %s", fqn)
	c.data = RowEvents{}
	c.fireChanged()
}

// header return resource header.
func (*Container) header() Row {
	var hh Row
	return append(hh,
		"NAME",
		"IMAGE",
		"READY",
		"STATE",
		"RS",
		"LPROB",
		"RPROB",
		"CPU",
		"MEM",
		"RCPU",
		"RMEM",
		"AGE",
	)
}

// Fields retrieves displayable fields.
func (c *Container) fields(pod *v1.Pod, co v1.Container) Row {
	ff := make(Row, 0, containerCols)

	// mxs, _ := c.MetricsServer.FetchPodsMetrics(r.pod.Namespace)

	// var cpu, mem string
	// for _, mx := range mxs.Items {
	// 	if mx.Name != r.pod.Name {
	// 		continue
	// 	}
	// 	for _, co := range mx.Containers {
	// 		if co.Name != i.Name {
	// 			continue
	// 		}
	// 		cpu, mem = toRes(co.Usage)
	// 	}
	// }

	// rcpu, rmem := resources(i)

	var cs *v1.ContainerStatus
	for _, cos := range pod.Status.ContainerStatuses {
		if cos.Name != co.Name {
			continue
		}
		cs = &cos
	}

	if cs == nil {
		for _, cos := range pod.Status.InitContainerStatuses {
			if cos.Name != co.Name {
				continue
			}
			cs = &cos
		}
	}

	ready, state, restarts := "false", MissingValue, "0"
	if cs != nil {
		ready, state, restarts = boolToStr(cs.Ready), toState(cs.State), strconv.Itoa(int(cs.RestartCount))
	}

	cpu, mem, rcpu, rmem := "Z", "Z", "Z", "Z"

	return append(ff,
		co.Name,
		co.Image,
		ready,
		state,
		restarts,
		probe(co.LivenessProbe),
		probe(co.ReadinessProbe),
		cpu,
		mem,
		rcpu,
		rmem,
		toAge(pod.CreationTimestamp),
	)
}

func (c *Container) updateData(action watch.EventType, po *v1.Pod) {
	for _, co := range po.Spec.InitContainers {
		ff := c.fields(po, co)

		if re, ok := c.data[co.Name]; ok {
			re.Action = action
			re.Deltas = re.Fields
			re.Fields = ff
		} else {
			c.data[co.Name] = &RowEvent{
				Action: action,
				Deltas: make(Row, containerCols),
				Fields: ff,
			}
		}
	}
	for _, co := range po.Spec.Containers {
		ff := c.fields(po, co)
		if re, ok := c.data[co.Name]; ok {
			re.Action = action
			re.Deltas = re.Fields
			re.Fields = ff
		} else {
			c.data[co.Name] = &RowEvent{
				Action: action,
				Deltas: make(Row, containerCols),
				Fields: ff,
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func toState(s v1.ContainerState) string {
	switch {
	case s.Waiting != nil:
		if s.Waiting.Reason != "" {
			return s.Waiting.Reason
		}
		return "Waiting"

	case s.Terminated != nil:
		if s.Terminated.Reason != "" {
			return s.Terminated.Reason
		}
		return "Terminated"
	case s.Running != nil:
		return "Running"
	default:
		return MissingValue
	}
}

func toRes(r v1.ResourceList) (string, string) {
	cpu, mem := r[v1.ResourceCPU], r[v1.ResourceMemory]

	return ToMillicore(cpu.MilliValue()), ToMi(k8s.ToMB(mem.Value()))
}

func resources(c v1.Container) (cpu, mem string) {
	req, lim := c.Resources.Requests, c.Resources.Limits

	if len(req) == 0 {
		if len(lim) != 0 {
			return toRes(lim)
		}
	} else {
		return toRes(req)
	}

	return "0", "0"
}

func probe(p *v1.Probe) string {
	if p == nil {
		return "no"
	}

	return "yes"
}
