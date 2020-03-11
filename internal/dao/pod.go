package dao

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const defaultTimeout = 8 * time.Second

var (
	_ Accessor   = (*Pod)(nil)
	_ Nuker      = (*Pod)(nil)
	_ Loggable   = (*Pod)(nil)
	_ Controller = (*Pod)(nil)
)

// Pod represents a pod resource.
type Pod struct {
	Resource
}

// IsHappy check for happy deployments.
func (p *Pod) IsHappy(po v1.Pod) bool {
	for _, c := range po.Status.Conditions {
		if c.Status == v1.ConditionFalse {
			return false
		}
	}
	return true
}

// Get returns a resource instance if found, else an error.
func (p *Pod) Get(ctx context.Context, path string) (runtime.Object, error) {
	o, err := p.Resource.Get(ctx, path)
	if err != nil {
		return o, err
	}

	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expecting *unstructured.Unstructured but got `%T", o)
	}

	var pmx *mv1beta1.PodMetrics
	if withMx, ok := ctx.Value(internal.KeyWithMetrics).(bool); withMx || !ok {
		if pmx, err = client.DialMetrics(p.Client()).FetchPodMetrics(path); err != nil {
			log.Debug().Err(err).Msgf("No pod metrics")
		}
	}

	return &render.PodWithMetrics{Raw: u, MX: pmx}, nil
}

// List returns a collection of nodes.
func (p *Pod) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	sel, ok := ctx.Value(internal.KeyFields).(string)
	if !ok {
		return nil, fmt.Errorf("expecting a fieldSelector in context")
	}
	fsel, err := labels.ConvertSelectorToLabelsMap(sel)
	if err != nil {
		return nil, err
	}
	nodeName := fsel["spec.nodeName"]

	oo, err := p.Resource.List(ctx, ns)
	if err != nil {
		return oo, err
	}

	var pmx *mv1beta1.PodMetricsList
	if withMx, ok := ctx.Value(internal.KeyWithMetrics).(bool); withMx || !ok {
		if pmx, err = client.DialMetrics(p.Client()).FetchPodsMetrics(ns); err != nil {
			log.Debug().Err(err).Msgf("No pods metrics")
		}
	}

	var res []runtime.Object
	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			return res, fmt.Errorf("expecting *unstructured.Unstructured but got `%T", o)
		}
		if nodeName == "" {
			res = append(res, &render.PodWithMetrics{Raw: u, MX: podMetricsFor(o, pmx)})
			continue
		}

		spec, ok := u.Object["spec"].(map[string]interface{})
		if !ok {
			return res, fmt.Errorf("expecting interface map but got `%T", o)
		}
		if spec["nodeName"] == nodeName {
			res = append(res, &render.PodWithMetrics{Raw: u, MX: podMetricsFor(o, pmx)})
		}
	}

	return res, nil
}

// Logs fetch container logs for a given pod and container.
func (p *Pod) Logs(path string, opts *v1.PodLogOptions) (*restclient.Request, error) {
	ns, _ := client.Namespaced(path)
	auth, err := p.Client().CanI(ns, "v1/pods:log", []string{client.GetVerb})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("user is not authorized to view pod logs")
	}

	ns, n := client.Namespaced(path)
	return p.Client().DialOrDie().CoreV1().Pods(ns).GetLogs(n, opts), nil
}

// Containers returns all container names on pod
func (p *Pod) Containers(path string, includeInit bool) ([]string, error) {
	pod, err := p.GetInstance(path)
	if err != nil {
		return nil, err
	}

	cc := make([]string, 0, len(pod.Spec.Containers)+len(pod.Spec.InitContainers))
	for _, c := range pod.Spec.Containers {
		cc = append(cc, c.Name)
	}

	if includeInit {
		for _, c := range pod.Spec.InitContainers {
			cc = append(cc, c.Name)
		}
	}

	return cc, nil
}

// Pod returns a pod victim by name.
func (p *Pod) Pod(fqn string) (string, error) {
	return fqn, nil
}

// GetInstance returns a pod instance.
func (p *Pod) GetInstance(fqn string) (*v1.Pod, error) {
	o, err := p.Factory.Get(p.gvr.String(), fqn, false, labels.Everything())
	if err != nil {
		return nil, err
	}

	var pod v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
	if err != nil {
		return nil, err
	}

	return &pod, nil
}

// TailLogs tails a given container logs
func (p *Pod) TailLogs(ctx context.Context, c chan<- []byte, opts LogOptions) error {
	if !opts.HasContainer() {
		return p.logs(ctx, c, opts)
	}
	return tailLogs(ctx, p, c, opts)
}

func (p *Pod) logs(ctx context.Context, c chan<- []byte, opts LogOptions) error {
	fac, ok := ctx.Value(internal.KeyFactory).(*watch.Factory)
	if !ok {
		return errors.New("Expecting an informer")
	}
	o, err := fac.Get(p.gvr.String(), opts.Path, true, labels.Everything())
	if err != nil {
		return err
	}

	var po v1.Pod
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po); err != nil {
		return err
	}
	opts.Color = asColor(po.Name)
	if len(po.Spec.InitContainers)+len(po.Spec.Containers) == 1 {
		opts.SingleContainer = true
	}

	for _, co := range po.Spec.InitContainers {
		opts.Container = co.Name
		if err := p.TailLogs(ctx, c, opts); err != nil {
			return err
		}
	}
	rcos := loggableContainers(po.Status)
	for _, co := range po.Spec.Containers {
		if in(rcos, co.Name) {
			opts.Container = co.Name
			if err := p.TailLogs(ctx, c, opts); err != nil {
				log.Error().Err(err).Msgf("Getting logs for %s failed", co.Name)
				return err
			}
		}
	}

	return nil
}

func tailLogs(ctx context.Context, logger Logger, c chan<- []byte, opts LogOptions) error {
	log.Debug().Msgf("Tailing logs for %q -- %q", opts.Path, opts.Container)
	o := v1.PodLogOptions{
		Container:  opts.Container,
		Follow:     true,
		TailLines:  &opts.Lines,
		Timestamps: opts.ShowTimestamp,
		Previous:   opts.Previous,
	}
	req, err := logger.Logs(opts.Path, &o)
	if err != nil {
		return err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	req.Context(ctx)

	var blocked int32 = 1
	var emptyLogs int32 = 0

	logsCheckerCtx, logsCheckerCancel := context.WithCancel(ctx)
	go logsEmptyChecker(logsCheckerCtx, cancel, logger, opts, &blocked, &emptyLogs)
	go logsTimeout(cancel, &blocked)

	// This call will block if nothing is in the stream!!
	// That's why we doing parallel request above which will
	// cancel this stream in case logs size is zero
	// But we anyways using timeout in both contexts in case of connection issues
	stream, err := req.Stream()

	atomic.StoreInt32(&blocked, 0)
	logsCheckerCancel()

	if err != nil {
		log.Error().Err(err).Msgf("Log stream failed for `%s", opts.Path)

		if atomic.LoadInt32(&emptyLogs) != 0 {
			return fmt.Errorf("Empty log. Please reopen this view to see log changes.")
		} else {
			return fmt.Errorf("Unable to obtain log stream for %s", opts.Path)
		}
	}
	go readLogs(stream, c, opts)

	return nil
}

// Check logs size and in case its zero cancel blocked stream request
func logsEmptyChecker(ctx context.Context, cancel context.CancelFunc, logger Logger, opts LogOptions, blocked *int32, emptyLogs *int32) {
	var bytesLimit int64 = 1
	o := v1.PodLogOptions{
		Container:  opts.Container,
		Previous:   opts.Previous,
		LimitBytes: &bytesLimit,
	}

	req, err := logger.Logs(opts.Path, &o)
	if err != nil {
		log.Error().Err(err).Msgf("Log stream empty checker request build")
	}
	req.Context(ctx)

	buf, err := req.DoRaw()
	if err != nil {
		log.Error().Err(err).Msgf("Log stream empty checker request fail")
	}

	if atomic.LoadInt32(blocked) != 0 {
		if len(buf) == 0 {
			atomic.StoreInt32(emptyLogs, 1)
			log.Debug().Msg("Logs empty checker detected empty log")
			cancel()
		}
	}
}

func logsTimeout(cancel context.CancelFunc, blocked *int32) {
	<-time.After(defaultTimeout)
	if atomic.LoadInt32(blocked) == 1 {
		log.Debug().Msg("Timed out reading the log stream")
		cancel()
	}
}

func readLogs(stream io.ReadCloser, c chan<- []byte, opts LogOptions) {
	defer func() {
		log.Debug().Msgf(">>> Closing stream `%s", opts.Path)
		if err := stream.Close(); err != nil {
			log.Error().Err(err).Msg("Cloing stream")
		}
	}()

	r := bufio.NewReader(stream)
	for {
		bytes, err := r.ReadBytes('\n')
		if err != nil {
			log.Warn().Err(err).Msg("Read error")
			if err != io.EOF {
				log.Error().Err(err).Msgf("stream reader failed")
			}
			return
		}
		c <- opts.DecorateLog(bytes)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func podMetricsFor(o runtime.Object, mmx *mv1beta1.PodMetricsList) *mv1beta1.PodMetrics {
	if mmx == nil {
		return nil
	}
	fqn := extractFQN(o)
	for _, mx := range mmx.Items {
		if MetaFQN(mx.ObjectMeta) == fqn {
			return &mx
		}
	}
	return nil
}

// MetaFQN returns a fully qualified resource name.
func MetaFQN(m metav1.ObjectMeta) string {
	if m.Namespace == "" {
		return m.Name
	}

	return FQN(m.Namespace, m.Name)
}

// FQN returns a fully qualified resource name.
func FQN(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}

func extractFQN(o runtime.Object) string {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		log.Error().Err(fmt.Errorf("expecting unstructured but got %T", o))
		return "na"
	}
	m, ok := u.Object["metadata"].(map[string]interface{})
	if !ok {
		log.Error().Err(fmt.Errorf("expecting interface map for metadata but got %T", u.Object["metadata"]))
		return "na"
	}

	n, ok := m["name"].(string)
	if !ok {
		log.Error().Err(fmt.Errorf("expecting interface map for name but got %T", m["name"]))
		return "na"
	}

	ns, ok := m["namespace"].(string)
	if !ok {
		return FQN("", n)
	}

	return FQN(ns, n)
}

func loggableContainers(s v1.PodStatus) []string {
	var rcos []string
	for _, c := range s.ContainerStatuses {
		rcos = append(rcos, c.Name)
	}
	return rcos
}

func asColor(n string) color.Paint {
	var sum int
	for _, r := range n {
		sum += int(r)
	}
	return color.Paint(30 + 2 + sum%6)
}

// Check if string is in a string list.
func in(ll []string, s string) bool {
	for _, l := range ll {
		if l == s {
			return true
		}
	}
	return false
}
