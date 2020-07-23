package dao

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

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

const (
	logRetryCount = 20
	logRetryWait  = 1 * time.Second
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
		sel = ""
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

	dial, err := p.Client().Dial()
	if err != nil {
		return nil, err
	}

	ns, n := client.Namespaced(path)
	return dial.CoreV1().Pods(ns).GetLogs(n, opts), nil
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
	o, err := p.Factory.Get(p.gvr.String(), fqn, true, labels.Everything())
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
		if err := tailLogs(ctx, p, c, opts); err != nil {
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

// ScanSA scans for serviceaccount refs.
func (p *Pod) ScanSA(ctx context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := p.Factory.List(p.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var pod v1.Pod
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
		if err != nil {
			return nil, errors.New("expecting Deployment resource")
		}
		// Just pick controller less pods...
		if len(pod.ObjectMeta.OwnerReferences) > 0 {
			continue
		}
		if pod.Spec.ServiceAccountName == n {
			refs = append(refs, Ref{
				GVR: p.GVR(),
				FQN: client.FQN(pod.Namespace, pod.Name),
			})
		}
	}

	return refs, nil
}

// Scan scans for cluster resource refs.
func (p *Pod) Scan(ctx context.Context, gvr, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := p.Factory.List(p.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var pod v1.Pod
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
		if err != nil {
			return nil, errors.New("expecting Pod resource")
		}
		// Just pick controller less pods...
		if len(pod.ObjectMeta.OwnerReferences) > 0 {
			continue
		}
		switch gvr {
		case "v1/configmaps":
			if !hasConfigMap(&pod.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: p.GVR(),
				FQN: client.FQN(pod.Namespace, pod.Name),
			})
		case "v1/secrets":
			found, err := hasSecret(p.Factory, &pod.Spec, pod.Namespace, n, wait)
			if err != nil {
				log.Warn().Err(err).Msgf("locate secret %q", fqn)
				continue
			}
			if !found {
				continue
			}
			refs = append(refs, Ref{
				GVR: p.GVR(),
				FQN: client.FQN(pod.Namespace, pod.Name),
			})
		case "v1/persistentvolumeclaims":
			if !hasPVC(&pod.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: p.GVR(),
				FQN: client.FQN(pod.Namespace, pod.Name),
			})
		}
	}

	return refs, nil
}

// ----------------------------------------------------------------------------
// Helpers...

func tailLogs(ctx context.Context, logger Logger, c LogChan, opts LogOptions) error {
	log.Debug().Msgf("Tailing logs for %q:%q", opts.Path, opts.Container)

	var (
		err    error
		req    *restclient.Request
		stream io.ReadCloser
	)
done:
	for r := 0; r < logRetryCount; r++ {
		log.Debug().Msgf("Retry logs %d", r)
		req, err = logger.Logs(opts.Path, opts.ToPodLogOptions())
		if err == nil {
			// This call will block if nothing is in the stream!!
			if stream, err = req.Stream(ctx); err == nil {
				log.Debug().Msgf("Reading logs")
				go readLogs(stream, c, opts)
				break
			} else {
				log.Error().Err(err).Msg("Streaming logs")
			}
		} else {
			log.Error().Err(err).Msg("Requesting logs")
		}

		select {
		case <-ctx.Done():
			err = ctx.Err()
			break done
		default:
			time.Sleep(logRetryWait)
		}
	}

	if err != nil {
		log.Error().Err(err).Msgf("Unable to obtain log stream failed for `%s", opts.Path)
		c <- opts.DecorateLog([]byte("\n" + err.Error() + "\n"))
		return err
	}

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
				c <- opts.DecorateLog([]byte("\nlog stream closed\n"))
				return
			}
			log.Warn().Err(err).Msgf("Stream READ error %s", opts.Info())
			c <- opts.DecorateLog([]byte(fmt.Sprintf("\nlog stream failed: %#v\n", err)))
			return
		}
		c <- opts.DecorateLog(bytes)
	}
}

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
