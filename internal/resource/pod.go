package resource

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/watch"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	defaultTimeout = 1 * time.Second
)

type (
	// IKey informer context key.
	IKey string

	// Containers represents a resource that supports containers.
	Containers interface {
		Containers(path string, includeInit bool) ([]string, error)
	}

	// Tailable represents a resource with tailable logs.
	Tailable interface {
		Logs(ctx context.Context, c chan<- string, opts LogOptions) error
	}

	// TailableResource is a resource that have tailable logs.
	TailableResource interface {
		Resource
		Tailable
	}

	// Pod that can be displayed in a table and interacted with.
	Pod struct {
		*Base
		instance *v1.Pod
		metrics  *mv1beta1.PodMetrics
	}
)

// NewPodList returns a new resource list.
func NewPodList(c Connection, ns string) List {
	return NewList(
		ns,
		"po",
		NewPod(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewPod instantiates a new Pod.
func NewPod(c Connection) *Pod {
	p := &Pod{
		Base: &Base{Connection: c, Resource: k8s.NewPod(c)},
	}
	p.Factory = p

	return p
}

// New builds a new Pod instance from a k8s resource.
func (r *Pod) New(i interface{}) Columnar {
	c := NewPod(r.Connection)
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

// SetPodMetrics set the current k8s resource metrics on a given pod.
func (r *Pod) SetPodMetrics(m *mv1beta1.PodMetrics) {
	r.metrics = m
}

// Marshal resource to yaml.
func (r *Pod) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}
	po := i.(*v1.Pod)
	po.TypeMeta.APIVersion = "v1"
	po.TypeMeta.Kind = "Pod"

	return r.marshalObject(po)
}

// Containers lists out all the docker containers name contained in a pod.
func (r *Pod) Containers(path string, includeInit bool) ([]string, error) {
	ns, po := Namespaced(path)

	return r.Resource.(k8s.Loggable).Containers(ns, po, includeInit)
}

// PodLogs tail logs for all containers in a running Pod.
func (r *Pod) PodLogs(ctx context.Context, c chan<- string, opts LogOptions) error {
	i := ctx.Value(IKey("informer")).(*watch.Informer)
	p, err := i.Get(watch.PodIndex, opts.FQN(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	po := p.(*v1.Pod)
	opts.Color = asColor(po.Name)
	if len(po.Spec.InitContainers)+len(po.Spec.Containers) == 1 {
		opts.SingleContainer = true
	}

	for _, co := range po.Spec.InitContainers {
		opts.Container = co.Name
		if err := r.Logs(ctx, c, opts); err != nil {
			return err
		}
	}
	rcos := r.loggableContainers(po.Status)
	for _, co := range po.Spec.Containers {
		if in(rcos, co.Name) {
			opts.Container = co.Name
			if err := r.Logs(ctx, c, opts); err != nil {
				log.Error().Err(err).Msgf("Getting logs for %s failed", co.Name)
				return err
			}
		}
	}

	return nil
}

// Logs tails a given container logs
func (r *Pod) Logs(ctx context.Context, c chan<- string, opts LogOptions) error {
	if !opts.HasContainer() {
		return r.PodLogs(ctx, c, opts)
	}
	res, ok := r.Resource.(k8s.Loggable)
	if !ok {
		return fmt.Errorf("Resource %T is not Loggable", r.Resource)
	}

	return tailLogs(ctx, res, c, opts)
}

func tailLogs(ctx context.Context, res k8s.Loggable, c chan<- string, opts LogOptions) error {
	log.Debug().Msgf("Tailing logs for %q/%q:%q", opts.Namespace, opts.Name, opts.Container)
	o := v1.PodLogOptions{
		Container: opts.Container,
		Follow:    true,
		TailLines: &opts.Lines,
		Previous:  opts.Previous,
	}
	req := res.Logs(opts.Namespace, opts.Name, &o)
	ctxt, cancelFunc := context.WithCancel(ctx)
	req.Context(ctxt)

	var blocked int32 = 1
	go logsTimeout(cancelFunc, &blocked)

	// This call will block if nothing is in the stream!!
	stream, err := req.Stream()
	atomic.StoreInt32(&blocked, 0)
	if err != nil {
		log.Error().Err(err).Msgf("Log stream failed for `%s", opts.Path())
		return fmt.Errorf("Unable to obtain log stream for %s", opts.Path())
	}
	go readLogs(ctx, stream, c, opts)

	return nil
}

func logsTimeout(cancel context.CancelFunc, blocked *int32) {
	select {
	case <-time.After(defaultTimeout):
		if atomic.LoadInt32(blocked) == 1 {
			log.Debug().Msg("Timed out reading the log stream")
			cancel()
		}
	}
}

func readLogs(ctx context.Context, stream io.ReadCloser, c chan<- string, opts LogOptions) {
	defer func() {
		log.Debug().Msgf(">>> Closing stream `%s", opts.Path())
		stream.Close()
	}()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			c <- opts.DecorateLog(scanner.Text())
		}
	}
}

// List resources for a given namespace.
func (r *Pod) List(ns string) (Columnars, error) {
	pods, err := r.Resource.List(ns)
	if err != nil {
		return nil, err
	}

	cc := make(Columnars, 0, len(pods))
	for i := range pods {
		po := r.New(&pods[i]).(*Pod)
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
		"%CPU",
		"%MEM",
		"IP",
		"NODE",
		"QOS",
		"AGE",
	)
}

// NumCols designates if column is numerical.
func (*Pod) NumCols(n string) map[string]bool {
	return map[string]bool{
		"CPU":  true,
		"MEM":  true,
		"%CPU": true,
		"%MEM": true,
		"RS":   true,
	}
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

	c, p := r.gatherPodMX(i)

	return append(ff,
		i.ObjectMeta.Name,
		strconv.Itoa(cr)+"/"+strconv.Itoa(len(ss)),
		r.phase(i),
		strconv.Itoa(rc),
		c.cpu,
		c.mem,
		p.cpu,
		p.mem,
		na(i.Status.PodIP),
		na(i.Spec.NodeName),
		r.mapQOS(i.Status.QOSClass),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func (r *Pod) gatherPodMX(po *v1.Pod) (c, p metric) {
	c, p = noMetric(), noMetric()
	if r.metrics == nil {
		return
	}

	cpu, mem := r.currentRes(r.metrics)
	c = metric{
		cpu: ToMillicore(cpu.MilliValue()),
		mem: ToMi(k8s.ToMB(mem.Value())),
	}

	rc, rm := r.requestedRes(po)
	p = metric{
		cpu: AsPerc(toPerc(float64(cpu.MilliValue()), float64(rc.MilliValue()))),
		mem: AsPerc(toPerc(k8s.ToMB(mem.Value()), k8s.ToMB(rm.Value()))),
	}

	return
}

func containerResources(co v1.Container) (cpu, mem *resource.Quantity) {
	req, limit := co.Resources.Requests, co.Resources.Limits
	switch {
	case len(req) != 0 && len(limit) != 0:
		cpu, mem = limit.Cpu(), limit.Memory()
	case len(req) != 0:
		cpu, mem = req.Cpu(), req.Memory()
	case len(limit) != 0:
		cpu, mem = limit.Cpu(), limit.Memory()
	}
	return
}

func (r *Pod) requestedRes(po *v1.Pod) (cpu, mem resource.Quantity) {
	for _, co := range po.Spec.Containers {
		c, m := containerResources(co)
		if c != nil {
			cpu.Add(*c)
		}
		if m != nil {
			mem.Add(*m)
		}
	}
	return
}

func (*Pod) currentRes(mx *mv1beta1.PodMetrics) (cpu, mem resource.Quantity) {
	for _, co := range mx.Containers {
		c, m := co.Usage.Cpu(), co.Usage.Memory()
		cpu.Add(*c)
		mem.Add(*m)
	}
	return
}

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
		if po.DeletionTimestamp != nil && po.Status.Reason == "NodeLost" {
			return "Unknown"
		}
		status = po.Status.Reason
	}

	init, status := r.initContainerPhase(po.Status, len(po.Spec.InitContainers), status)
	if init {
		return status
	}

	running, status := r.containerPhase(po.Status, status)
	if running && status == "Completed" {
		status = "Running"
	}
	if po.DeletionTimestamp == nil {
		return status
	}

	return "Terminating"
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
				status = "Signal:" + strconv.Itoa(int(cs.State.Terminated.Signal))
			} else {
				status = "ExitCode:" + strconv.Itoa(int(cs.State.Terminated.ExitCode))
			}
		case cs.Ready && cs.State.Running != nil:
			running = true
		}
	}

	return running, status
}

func (*Pod) initContainerPhase(st v1.PodStatus, initCount int, status string) (bool, string) {
	for i, cs := range st.InitContainerStatuses {
		status := checkContainerStatus(cs, i, initCount)
		if status == "" {
			continue
		}
		return true, status
	}

	return false, status
}

func checkContainerStatus(cs v1.ContainerStatus, i, initCount int) string {
	switch {
	case cs.State.Terminated != nil:
		if cs.State.Terminated.ExitCode == 0 {
			return ""
		}
		if cs.State.Terminated.Reason != "" {
			return "Init:" + cs.State.Terminated.Reason
		}
		if cs.State.Terminated.Signal != 0 {
			return "Init:Signal:" + strconv.Itoa(int(cs.State.Terminated.Signal))
		}
		return "Init:ExitCode:" + strconv.Itoa(int(cs.State.Terminated.ExitCode))
	case cs.State.Waiting != nil && cs.State.Waiting.Reason != "" && cs.State.Waiting.Reason != "PodInitializing":
		return "Init:" + cs.State.Waiting.Reason
	default:
		return "Init:" + strconv.Itoa(i) + "/" + strconv.Itoa(initCount)
	}
}

func (r *Pod) loggableContainers(s v1.PodStatus) []string {
	var rcos []string
	for _, c := range s.ContainerStatuses {
		rcos = append(rcos, c.Name)
	}
	return rcos
}

// Helpers..

func asColor(n string) color.Paint {
	var sum int
	for _, r := range n {
		sum += int(r)
	}
	return color.Paint(30 + 2 + sum%6)
}
