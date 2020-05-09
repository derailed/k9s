package dao

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
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
		if pmx, err = client.DialMetrics(p.Client()).FetchPodMetrics(ctx, path); err != nil {
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
		if pmx, err = client.DialMetrics(p.Client()).FetchPodsMetrics(ctx, ns); err != nil {
			log.Debug().Err(err).Msgf("No pods metrics")
		}
	}

	res := make([]runtime.Object, 0, len(oo))
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
func (p *Pod) TailLogs(ctx context.Context, c LogChan, opts LogOptions) error {
	log.Debug().Msgf("TAIL-LOGS for %q:%q", opts.Path, opts.Container)
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

	if opts.HasContainer() {
		opts.SingleContainer = true
		return tailLogs(ctx, p, c, opts)
	}
	if len(po.Spec.InitContainers)+len(po.Spec.Containers) == 1 {
		opts.SingleContainer = true
	}

	var tailed bool
	for _, co := range po.Spec.InitContainers {
		log.Debug().Msgf("Tailing INIT-CO %q", co.Name)
		opts.Container = co.Name
		if err := p.TailLogs(ctx, c, opts); err != nil {
			return err
		}
		tailed = true
	}
	for _, co := range po.Spec.Containers {
		log.Debug().Msgf("Tailing CO %q", co.Name)
		opts.Container = co.Name
		if err := tailLogs(ctx, p, c, opts); err != nil {
			return err
		}
		tailed = true
	}
	for _, co := range po.Spec.EphemeralContainers {
		log.Debug().Msgf("Tailing EPH-CO %q", co.Name)
		opts.Container = co.Name
		if err := tailLogs(ctx, p, c, opts); err != nil {
			return err
		}
		tailed = true
	}

	if !tailed {
		return fmt.Errorf("no loggable containers found for pod %s", opts.Path)
	}

	return nil
}

func tailLogs(ctx context.Context, logger Logger, c LogChan, opts LogOptions) error {
	log.Debug().Msgf("Tailing logs for %q:%q", opts.Path, opts.Container)
	req, err := logger.Logs(opts.Path, opts.ToPodLogOptions())
	if err != nil {
		return err
	}

	// This call will block if nothing is in the stream!!
	stream, err := req.Stream(ctx)
	if err != nil {
		c <- opts.DecorateLog([]byte(err.Error() + "\n"))
		log.Error().Err(err).Msgf("Unable to obtain log stream failed for `%s", opts.Path)
		return err
	}
	go readLogs(stream, c, opts)

	return nil
}

func readLogs(stream io.ReadCloser, c LogChan, opts LogOptions) {
	defer func() {
		log.Debug().Msgf(">>> Closing stream %s", opts.Info())
		if err := stream.Close(); err != nil {
			log.Error().Err(err).Msgf("Fail to close stream %s", opts.Info())
		}
	}()

	r := bufio.NewReader(stream)
	for {
		bytes, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Warn().Err(err).Msgf("Stream closed for %s", opts.Info())
				c <- opts.DecorateLog([]byte("log stream closed\n"))
				return
			}
			log.Warn().Err(err).Msgf("Stream READ error %s", opts.Info())
			c <- opts.DecorateLog([]byte("log stream failed\n"))
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

// Check if string is in a string list.
func in(ll []string, s string) bool {
	for _, l := range ll {
		if l == s {
			return true
		}
	}
	return false
}
