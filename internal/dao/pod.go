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
	"k8s.io/apimachinery/pkg/types"
	restclient "k8s.io/client-go/rest"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var (
	_ Accessor        = (*Pod)(nil)
	_ Nuker           = (*Pod)(nil)
	_ Loggable        = (*Pod)(nil)
	_ Controller      = (*Pod)(nil)
	_ ContainsPodSpec = (*Pod)(nil)
)

const (
	logRetryCount                 = 20
	logRetryWait                  = 1 * time.Second
	defaultLogContainerAnnotation = "kubectl.kubernetes.io/default-logs-container"
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
		pmx, _ = client.DialMetrics(p.Client()).FetchPodMetrics(ctx, path)
	}

	return &render.PodWithMetrics{Raw: u, MX: pmx}, nil
}

// List returns a collection of nodes.
func (p *Pod) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	oo, err := p.Resource.List(ctx, ns)
	if err != nil {
		return oo, err
	}

	var pmx client.PodsMetricsMap
	if withMx, ok := ctx.Value(internal.KeyWithMetrics).(bool); withMx || !ok {
		pmx, _ = client.DialMetrics(p.Client()).FetchPodsMetricsMap(ctx, ns)
	}
	sel, _ := ctx.Value(internal.KeyFields).(string)
	fsel, err := labels.ConvertSelectorToLabelsMap(sel)
	if err != nil {
		return nil, err
	}
	nodeName := fsel["spec.nodeName"]

	res := make([]runtime.Object, 0, len(oo))
	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			return res, fmt.Errorf("expecting *unstructured.Unstructured but got `%T", o)
		}
		fqn := extractFQN(o)
		if nodeName == "" {
			res = append(res, &render.PodWithMetrics{Raw: u, MX: pmx[fqn]})
			continue
		}

		spec, ok := u.Object["spec"].(map[string]interface{})
		if !ok {
			return res, fmt.Errorf("expecting interface map but got `%T", o)
		}
		if spec["nodeName"] == nodeName {
			res = append(res, &render.PodWithMetrics{Raw: u, MX: pmx[fqn]})
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

// Containers returns all container names on pod.
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

// TailLogs tails a given container logs.
func (p *Pod) TailLogs(ctx context.Context, out LogChan, opts *LogOptions) error {
	log.Debug().Msgf("TAIL-LOGS for %q:%q", opts.Path, opts.Container)
	fac, ok := ctx.Value(internal.KeyFactory).(*watch.Factory)
	if !ok {
		return errors.New("No factory in context")
	}
	o, err := fac.Get(p.gvr.String(), opts.Path, true, labels.Everything())
	if err != nil {
		return err
	}
	var po v1.Pod
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po); err != nil {
		return err
	}

	if len(po.Spec.InitContainers)+len(po.Spec.Containers) == 1 {
		opts.SingleContainer = true
	}

	if co, ok := GetDefaultLogContainer(po.ObjectMeta, po.Spec); ok && !opts.AllContainers {
		opts.DefaultContainer = co
		return tailLogs(ctx, p, out, opts)
	}

	if opts.HasContainer() && !opts.AllContainers {
		return tailLogs(ctx, p, out, opts)
	}

	var tailed bool
	for _, co := range po.Spec.InitContainers {
		o := opts.Clone()
		o.Container = co.Name
		if err := tailLogs(ctx, p, out, o); err != nil {
			return err
		}
		tailed = true
	}
	for _, co := range po.Spec.Containers {
		o := opts.Clone()
		o.Container = co.Name
		if err := tailLogs(ctx, p, out, o); err != nil {
			return err
		}
		tailed = true
	}
	for _, co := range po.Spec.EphemeralContainers {
		o := opts.Clone()
		o.Container = co.Name
		if err := tailLogs(ctx, p, out, o); err != nil {
			return err
		}
		tailed = true
	}

	if !tailed {
		return fmt.Errorf("no loggable containers found for pod %s", opts.Path)
	}

	return nil
}

// ScanSA scans for ServiceAccount refs.
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

func tailLogs(ctx context.Context, logger Logger, out LogChan, opts *LogOptions) error {
	var (
		err    error
		req    *restclient.Request
		stream io.ReadCloser
	)

	o := opts.ToPodLogOptions()
	log.Debug().Msgf("TAIL_LOGS! %#v", o)
done:
	for r := 0; r < logRetryCount; r++ {
		req, err = logger.Logs(opts.Path, o)
		if err == nil {
			// This call will block if nothing is in the stream!!
			if stream, err = req.Stream(ctx); err == nil {
				go readLogs(ctx, stream, out, opts)
				break
			} else {
				log.Error().Err(err).Msg("Streaming logs")
			}
		} else {
			log.Error().Err(err).Msg("Requesting logs")
		}

		select {
		case <-ctx.Done():
			log.Debug().Msgf("!!!!TAIL_LOGS CANCELED!!!!")
			err = ctx.Err()
			break done
		default:
			time.Sleep(logRetryWait)
		}
	}

	return err
}

func readLogs(ctx context.Context, stream io.ReadCloser, c LogChan, opts *LogOptions) {
	defer func() {
		log.Debug().Msgf("READ_LOGS BAILED!!!")
		if err := stream.Close(); err != nil {
			log.Error().Err(err).Msgf("Fail to close stream %s", opts.Info())
		}
	}()

	log.Debug().Msgf("READ_LOGS PROCESSING %#v", opts)
	r := bufio.NewReader(stream)
	for {
		bytes, err := r.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Warn().Err(err).Msgf("Stream closed for %s", opts.Info())
				// c <- ItemEOF
				return
			}
			log.Warn().Err(err).Msgf("Stream READ error %s", opts.Info())
			return
		}
		select {
		case c <- opts.DecorateLog(bytes):
		case <-ctx.Done():
			log.Debug().Msgf("READER CANCELED")
			return
		}
	}
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
		return client.NA
	}
	m, ok := u.Object["metadata"].(map[string]interface{})
	if !ok {
		log.Error().Err(fmt.Errorf("expecting interface map for metadata but got %T", u.Object["metadata"]))
		return client.NA
	}

	n, ok := m["name"].(string)
	if !ok {
		log.Error().Err(fmt.Errorf("expecting interface map for name but got %T", m["name"]))
		return client.NA
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

// GetPodSpec returns a pod spec given a resource.
func (p *Pod) GetPodSpec(path string) (*v1.PodSpec, error) {
	pod, err := p.GetInstance(path)
	if err != nil {
		return nil, err
	}
	podSpec := pod.Spec
	return &podSpec, nil
}

// SetImages sets container images.
func (p *Pod) SetImages(ctx context.Context, path string, imageSpecs ImageSpecs) error {
	ns, n := client.Namespaced(path)
	auth, err := p.Client().CanI(ns, "v1/pod", []string{client.PatchVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to patch a deployment")
	}
	manager, isManaged, err := p.isControlled(path)
	if err != nil {
		return err
	}
	if isManaged {
		return fmt.Errorf("Unable to set image. This pod is managed by %s. Please set the image on the controller", manager)
	}
	jsonPatch, err := GetJsonPatch(imageSpecs)
	if err != nil {
		return err
	}
	dial, err := p.Client().Dial()
	if err != nil {
		return err
	}
	_, err = dial.CoreV1().Pods(ns).Patch(
		ctx,
		n,
		types.StrategicMergePatchType,
		jsonPatch,
		metav1.PatchOptions{},
	)
	return err
}

func (p *Pod) isControlled(path string) (string, bool, error) {
	pod, err := p.GetInstance(path)
	if err != nil {
		return "", false, err
	}
	references := pod.GetObjectMeta().GetOwnerReferences()
	if len(references) > 0 {
		return fmt.Sprintf("%s/%s", references[0].Kind, references[0].Name), true, nil
	}
	return "", false, nil
}

// GetDefaultLogContainer returns a container name if specified in an annotation.
func GetDefaultLogContainer(m metav1.ObjectMeta, spec v1.PodSpec) (string, bool) {
	defaultContainer, ok := m.Annotations[defaultLogContainerAnnotation]
	if ok {
		for _, container := range spec.Containers {
			if container.Name == defaultContainer {
				return defaultContainer, true
			}
		}
		log.Warn().Msg(defaultContainer + " container  not found. " + defaultLogContainerAnnotation + " annotation will be ignored")
	}

	return "", false
}
