package resource

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/util/node"
)

const (
	defaultTimeout = 1 * time.Second
)

type (
	// Containers represents a resource that supports containers.
	Containers interface {
		Containers(path string, includeInit bool) ([]string, error)
	}

	// Tailable represents a resource with tailable logs.
	Tailable interface {
		Logs(c chan<- string, ns, na, co string, lines int64, prev bool) (context.CancelFunc, error)
	}

	// TailableResource is a resource that have tailable logs.
	TailableResource interface {
		Resource
		Tailable
	}

	// Pod that can be displayed in a table and interacted with.
	Pod struct {
		*Base
		instance      *v1.Pod
		MetricsServer MetricsServer
		metrics       k8s.PodMetrics
		mx            sync.RWMutex
	}
)

// NewPodList returns a new resource list.
func NewPodList(c Connection, mx MetricsServer, ns string) List {
	return NewList(
		ns,
		"po",
		NewPod(c, mx),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewPod instantiates a new Pod.
func NewPod(c Connection, mx MetricsServer) *Pod {
	p := &Pod{
		Base:          &Base{Connection: c, Resource: k8s.NewPod(c)},
		MetricsServer: mx,
		metrics:       k8s.PodMetrics{},
	}
	p.Factory = p

	return p
}

// New builds a new Pod instance from a k8s resource.
func (r *Pod) New(i interface{}) Columnar {
	c := NewPod(r.Connection, r.MetricsServer)
	switch instance := i.(type) {
	case *v1.Pod:
		c.instance = instance
	case v1.Pod:
		c.instance = &instance
	case *interface{}:
		ptr := *instance
		po := ptr.(v1.Pod)
		c.instance = &po
	default:
		log.Fatal().Msgf("unknown Pod type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Metrics retrieves cpu/mem resource consumption on a pod.
func (r *Pod) Metrics() k8s.PodMetrics {
	return r.metrics
}

// SetMetrics set the current k8s resource metrics on a given pod.
func (r *Pod) SetMetrics(m k8s.PodMetrics) {
	r.metrics = m
}

// Marshal resource to yaml.
func (r *Pod) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	po := i.(*v1.Pod)
	po.TypeMeta.APIVersion = "v1"
	po.TypeMeta.Kind = "Pod"

	return r.marshalObject(po)
}

// Containers lists out all the docker contrainers name contained in a pod.
func (r *Pod) Containers(path string, includeInit bool) ([]string, error) {
	ns, po := namespaced(path)

	return r.Resource.(k8s.Loggable).Containers(ns, po, includeInit)
}

// Logs tails a given container logs
func (r *Pod) Logs(c chan<- string, ns, n, co string, lines int64, prev bool) (context.CancelFunc, error) {
	req := r.Resource.(k8s.Loggable).Logs(ns, n, co, lines, prev)
	ctx, cancel := context.WithCancel(context.TODO())
	req.Context(ctx)

	blocked := true
	go func() {
		select {
		case <-time.After(defaultTimeout):
			var closes bool
			r.mx.RLock()
			{
				closes = blocked
			}
			r.mx.RUnlock()
			if closes {
				log.Debug().Msg(">>Closing Channel<<")
				close(c)
				cancel()
			}
		}
	}()

	// This call will block if nothing is in the stream!!
	stream, err := req.Stream()
	if err != nil {
		log.Error().Msgf("Tail logs failed `%s/%s:%s -- %v", ns, n, co, err)
		return cancel, err
	}

	r.mx.Lock()
	{
		blocked = false
	}
	r.mx.Unlock()

	go func() {
		defer func() {
			log.Debug().Msg("!!!Closing Stream!!!")
			close(c)
			stream.Close()
			cancel()
		}()

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			c <- scanner.Text()
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	return cancel, nil
}

// List resources for a given namespace.
func (r *Pod) List(ns string) (Columnars, error) {
	pods, err := r.Resource.List(ns)
	if err != nil {
		return nil, err
	}

	mx := make(k8s.PodsMetrics, len(pods))
	if r.MetricsServer.HasMetrics() {
		pmx, _ := r.MetricsServer.FetchPodsMetrics(ns)
		r.MetricsServer.PodsMetrics(pmx, mx)
	}

	cc := make(Columnars, 0, len(pods))
	for i := range pods {
		po := r.New(&pods[i]).(*Pod)
		if err == nil {
			po.metrics = mx[po.Name()]
		}
		cc = append(cc, po)
	}

	return cc, nil
}

// Header return resource header.
func (*Pod) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	return append(hh,
		"NAME",
		"READY",
		"STATUS",
		"RS",
		"CPU",
		"MEM",
		"IP",
		"NODE",
		"QOS",
		"AGE",
	)
}

// Fields retrieves displayable fields.
func (r *Pod) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance

	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	ss := i.Status.ContainerStatuses
	cr, _, rc := r.statuses(ss)

	return append(ff,
		i.ObjectMeta.Name,
		strconv.Itoa(cr)+"/"+strconv.Itoa(len(ss)),
		r.phase(i),
		strconv.Itoa(rc),
		ToMillicore(r.metrics.CurrentCPU),
		ToMi(r.metrics.CurrentMEM),
		i.Status.PodIP,
		i.Spec.NodeName,
		r.mapQOS(i.Status.QOSClass),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func (*Pod) mapQOS(class v1.PodQOSClass) string {
	switch class {
	case v1.PodQOSGuaranteed:
		return "GA"
	case v1.PodQOSBurstable:
		return "BU"
	default:
		return "BE"
	}
}

func (r *Pod) statuses(ss []v1.ContainerStatus) (cr, ct, rc int) {
	for _, c := range ss {
		if c.State.Terminated != nil {
			ct++
		}
		if c.Ready {
			cr = cr + 1
		}
		rc += int(c.RestartCount)
	}

	return
}

func isSet(s *string) bool {
	return s != nil && *s != ""
}

func (r *Pod) phase(po *v1.Pod) string {
	status := string(po.Status.Phase)
	if po.Status.Reason != "" {
		if po.DeletionTimestamp != nil && po.Status.Reason == node.NodeUnreachablePodReason {
			return "Unknown"
		}
		status = po.Status.Reason
	}

	var init bool
	init, status = r.initContainerPhase(po.Status, len(po.Spec.InitContainers), status)
	if init {
		return status
	}

	var running bool
	running, status = r.containerPhase(po.Status, status)
	if status == "Completed" && running {
		status = "Running"
	}

	if po.DeletionTimestamp == nil {
		return status
	}

	return "Terminated"
}

func (*Pod) containerPhase(st v1.PodStatus, status string) (bool, string) {
	var running bool
	for i := len(st.ContainerStatuses) - 1; i >= 0; i-- {
		cs := st.ContainerStatuses[i]
		switch {
		case cs.State.Waiting != nil && cs.State.Waiting.Reason != "":
			status = cs.State.Waiting.Reason
		case cs.State.Terminated != nil && cs.State.Terminated.Reason != "":
			status = cs.State.Terminated.Reason
		case cs.State.Terminated != nil:
			if cs.State.Terminated.Signal != 0 {
				status = fmt.Sprintf("Signal:%d", cs.State.Terminated.Signal)
			} else {
				status = fmt.Sprintf("ExitCode:%d", cs.State.Terminated.ExitCode)
			}
		case cs.Ready && cs.State.Running != nil:
			running = true
		}
	}

	return running, status
}

func (*Pod) initContainerPhase(st v1.PodStatus, initCount int, status string) (bool, string) {
	var init bool
	for i, cs := range st.InitContainerStatuses {
		switch {
		case cs.State.Terminated != nil:
			if cs.State.Terminated.ExitCode == 0 {
				continue
			}
			if cs.State.Terminated.Reason != "" {
				status = "Init:" + cs.State.Terminated.Reason
				break
			}
			if cs.State.Terminated.Signal != 0 {
				status = fmt.Sprintf("Init:Signal:%d", cs.State.Terminated.Signal)
			} else {
				status = fmt.Sprintf("Init:ExitCode:%d", cs.State.Terminated.ExitCode)
			}
		case cs.State.Waiting != nil && cs.State.Waiting.Reason != "" && cs.State.Waiting.Reason != "PodInitializing":
			status = "Init:" + cs.State.Waiting.Reason
		default:
			status = fmt.Sprintf("Init:%d/%d", i, initCount)
		}
		init = true
		break
	}

	return init, status
}
