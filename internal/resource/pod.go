package resource

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

const (
	defaultTimeout = 1 * time.Second
	podNameSize    = 42
)

type (
	// Container represents a resource that encompass multiple containers.
	Container interface {
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
		instance  *v1.Pod
		metricSvc MetricsIfc
		metrics   k8s.Metric
	}
)

// NewPodList returns a new resource list.
func NewPodList(ns string) List {
	return NewPodListWithArgs(ns, NewPod())
}

// NewPodListWithArgs returns a new resource list.
func NewPodListWithArgs(ns string, res Resource) List {
	l := newList(ns, "po", res, AllVerbsAccess|DescribeAccess)
	l.xray = true
	return l
}

// NewPod returns a new Pod instance.
func NewPod() *Pod {
	return NewPodWithArgs(k8s.NewPod(), k8s.NewMetricsServer())
}

// NewPodWithArgs returns a new Pod instance.
func NewPodWithArgs(r k8s.Res, mx MetricsIfc) *Pod {
	p := &Pod{
		metricSvc: mx,
		Base: &Base{
			caller: r,
		},
	}
	p.creator = p
	return p
}

// NewInstance builds a new Pod instance from a k8s resource.
func (r *Pod) NewInstance(i interface{}) Columnar {
	pod := NewPod()
	switch i.(type) {
	case *v1.Pod:
		pod.instance = i.(*v1.Pod)
	case v1.Pod:
		ii := i.(v1.Pod)
		pod.instance = &ii
	case *interface{}:
		ptr := *i.(*interface{})
		po := ptr.(v1.Pod)
		pod.instance = &po
	default:
		log.Fatalf("Unknown %#v", i)
	}
	pod.path = r.namespacedName(pod.instance.ObjectMeta)
	return pod
}

// Metrics retrieves cpu/mem resource consumption on a pod.
func (r *Pod) Metrics() k8s.Metric {
	return r.metrics
}

// SetMetrics set the current k8s resource metrics on a given pod.
func (r *Pod) SetMetrics(m k8s.Metric) {
	r.metrics = m
}

// Marshal resource to yaml.
func (r *Pod) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
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
	return r.caller.(k8s.Loggable).Containers(ns, po, includeInit)
}

// Logs tails a given container logs
func (r *Pod) Logs(c chan<- string, ns, n, co string, lines int64, prev bool) (context.CancelFunc, error) {
	req := r.caller.(k8s.Loggable).Logs(ns, n, co, lines, prev)
	ctx, cancel := context.WithCancel(context.TODO())
	req.Context(ctx)

	blocked := true
	go func() {
		select {
		case <-time.After(defaultTimeout):
			if blocked {
				close(c)
				cancel()
			}
		}
	}()
	// This call will block if nothing is in the stream!!
	stream, err := req.Stream()
	blocked = false
	if err != nil {
		return cancel, fmt.Errorf("Log tail request failed for pod `%s/%s:%s", ns, n, co)
	}

	go func() {
		defer func() {
			stream.Close()
			cancel()
			close(c)
		}()

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			c <- scanner.Text()
		}
	}()
	return cancel, nil
}

// List resources for a given namespace.
func (r *Pod) List(ns string) (Columnars, error) {
	ii, err := r.caller.List(ns)
	if err != nil {
		return nil, err
	}

	metrics, err := r.metricSvc.PodMetrics()
	if err != nil {
		log.Warn(err)
	}

	cc := make(Columnars, 0, len(ii))
	for i := 0; i < len(ii); i++ {
		po := r.NewInstance(&ii[i]).(MxColumnar)
		if err == nil {
			po.SetMetrics(metrics[po.Name()])
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
		"RESTARTS",
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

	cr, _, rc, cc := r.statuses()
	return append(ff,
		Pad(i.ObjectMeta.Name, podNameSize),
		strconv.Itoa(cr)+"/"+strconv.Itoa(len(cc)),
		r.phase(i.Status),
		strconv.Itoa(rc),
		r.metrics.CPU,
		r.metrics.Mem,
		i.Status.PodIP,
		i.Status.HostIP,
		string(i.Status.QOSClass),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extra info about the resource.
func (r *Pod) ExtFields() Properties {
	i := r.instance

	// po := k8s.Pod{}
	// e, err := po.Events(i.Namespace, i.Name)
	// if err != nil {
	// 	log.Error("Boom!", err)
	// }
	// if len(e.Items) > 0 {
	// 	log.Println("Events", ee.Items)
	// }

	return Properties{
		"Priority":        strconv.Itoa(int(*i.Spec.Priority)),
		"Priority Class":  missing(i.Spec.PriorityClassName),
		"Labels":          mapToStr(i.Labels),
		"Annotations":     mapToStr(i.ObjectMeta.Annotations),
		"Containers":      r.toContainers(i.Spec.Containers),
		"Init Containers": r.toContainers(i.Spec.InitContainers),
		"Node Selectors":  mapToStr(i.Spec.NodeSelector),
		"Volumes":         r.toVolumes(i.Spec.Volumes),
		// "Events":          r.toEvents(e),
	}
}

// func (r *Pod) toEvents(e *v1.EventList) []string {
// 	ss := make([]string, 0, len(e.Items)+1)
// 	for _, h := range([]string{"Type", "Reason", "From", "Message", "Age"}) {
// 		ss[0] = fmt.Printf("%10s %10s %20s %30s", )
// 	}
// 	for i, e := range e.Items {
// 		ss[i] = e.

// 	}
// 	return ss
// }

func (r *Pod) toVolumes(vv []v1.Volume) map[string]interface{} {
	m := make(map[string]interface{}, len(vv))
	for _, v := range vv {
		m[v.Name] = r.toVolume(v)
	}
	return m
}

func (r *Pod) toVolume(v v1.Volume) map[string]interface{} {
	switch {
	case v.Secret != nil:
		return map[string]interface{}{
			"Type":     "Secret",
			"Name":     v.Secret.SecretName,
			"Optional": r.boolPtrToStr(v.Secret.Optional),
		}
	case v.AWSElasticBlockStore != nil:
		return map[string]interface{}{
			"Type":      v.AWSElasticBlockStore.FSType,
			"VolumeID":  v.AWSElasticBlockStore.VolumeID,
			"Partition": strconv.Itoa(int(v.AWSElasticBlockStore.Partition)),
			"ReadOnly":  boolToStr(v.AWSElasticBlockStore.ReadOnly),
		}
	default:
		return map[string]interface{}{}
	}
}

func (r *Pod) toContainers(cc []v1.Container) map[string]interface{} {
	m := make(map[string]interface{}, len(cc))
	for _, c := range cc {
		m[c.Name] = map[string]interface{}{
			"Image":       c.Image,
			"Environment": r.toEnv(c.Env),
		}
	}
	return m
}

func (r *Pod) toEnv(ee []v1.EnvVar) []string {
	if len(ee) == 0 {
		return []string{MissingValue}
	}

	ss := make([]string, len(ee))
	for i, e := range ee {
		s := r.toEnvFrom(e.ValueFrom)
		if len(s) == 0 {
			ss[i] = e.Name + "=" + e.Value
		} else {
			ss[i] = e.Name + "=" + e.Value + "(" + s + ")"
		}
	}
	return ss
}

func (r *Pod) toEnvFrom(e *v1.EnvVarSource) string {
	if e == nil {
		return MissingValue
	}

	var s string
	switch {
	case e.ConfigMapKeyRef != nil:
		f := e.ConfigMapKeyRef
		s += f.Name + ":" + f.Key + "(" + r.boolPtrToStr(f.Optional) + ")"
	case e.FieldRef != nil:
		f := e.FieldRef
		s += f.FieldPath + ":" + f.APIVersion
	case e.SecretKeyRef != nil:
		f := e.SecretKeyRef
		s += f.Name + ":" + f.Key + "(" + r.boolPtrToStr(f.Optional) + ")"
	}
	return s
}

func (r *Pod) boolPtrToStr(b *bool) string {
	if b == nil {
		return "false"
	}

	return boolToStr(*b)
}

func (r *Pod) statuses() (cr, ct, rc int, cc []v1.ContainerStatus) {
	cc = r.instance.Status.ContainerStatuses
	for _, c := range cc {
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

func (*Pod) phase(s v1.PodStatus) string {
	status := "Pending"
	for _, cs := range s.ContainerStatuses {
		switch {
		case cs.State.Running != nil:
			status = "Running"
		case cs.State.Waiting != nil:
			status = cs.State.Waiting.Reason
		case cs.State.Terminated != nil:
			status = "Terminating"
			if len(cs.State.Terminated.Reason) != 0 {
				status = cs.State.Terminated.Reason
			}
		}
	}
	return status
}
