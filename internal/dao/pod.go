// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tview"
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
	_ ImageLister     = (*Pod)(nil)
)

const (
	logRetryCount = 20
	logRetryWait  = 1 * time.Second
)

// Pod represents a pod resource.
type Pod struct {
	Resource
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
	if withMx, ok := ctx.Value(internal.KeyWithMetrics).(bool); ok && withMx {
		pmx, _ = client.DialMetrics(p.Client()).FetchPodMetrics(ctx, path)
	}

	return &render.PodWithMetrics{Raw: u, MX: pmx}, nil
}

// ListImages lists container images.
func (p *Pod) ListImages(ctx context.Context, path string) ([]string, error) {
	pod, err := p.GetInstance(path)
	if err != nil {
		return nil, err
	}

	return render.ExtractImages(&pod.Spec), nil
}

// List returns a collection of nodes.
func (p *Pod) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	oo, err := p.Resource.List(ctx, ns)
	if err != nil {
		return oo, err
	}

	var pmx client.PodsMetricsMap
	if withMx, ok := ctx.Value(internal.KeyWithMetrics).(bool); ok && withMx {
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
	ns, n := client.Namespaced(path)
	auth, err := p.Client().CanI(ns, "v1/pods:log", n, client.GetAccess)
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("user is not authorized to view pod logs")
	}

	dial, err := p.Client().DialLogs()
	if err != nil {
		return nil, err
	}

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
	o, err := p.getFactory().Get(p.gvrStr(), fqn, true, labels.Everything())
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
func (p *Pod) TailLogs(ctx context.Context, opts *LogOptions) ([]LogChan, error) {
	fac, ok := ctx.Value(internal.KeyFactory).(*watch.Factory)
	if !ok {
		return nil, errors.New("no factory in context")
	}
	o, err := fac.Get(p.gvrStr(), opts.Path, true, labels.Everything())
	if err != nil {
		return nil, err
	}
	var po v1.Pod
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po); err != nil {
		return nil, err
	}
	coCounts := len(po.Spec.InitContainers) + len(po.Spec.Containers) + len(po.Spec.EphemeralContainers)
	if coCounts == 1 {
		opts.SingleContainer = true
	}

	outs := make([]LogChan, 0, coCounts)
	if co, ok := GetDefaultContainer(po.ObjectMeta, po.Spec); ok && !opts.AllContainers {
		opts.DefaultContainer = co
		return append(outs, tailLogs(ctx, p, opts)), nil
	}
	if opts.HasContainer() && !opts.AllContainers {
		return append(outs, tailLogs(ctx, p, opts)), nil
	}
	for _, co := range po.Spec.InitContainers {
		cfg := opts.Clone()
		cfg.Container = co.Name
		outs = append(outs, tailLogs(ctx, p, cfg))
	}
	for _, co := range po.Spec.Containers {
		cfg := opts.Clone()
		cfg.Container = co.Name
		outs = append(outs, tailLogs(ctx, p, cfg))
	}
	for _, co := range po.Spec.EphemeralContainers {
		cfg := opts.Clone()
		cfg.Container = co.Name
		outs = append(outs, tailLogs(ctx, p, cfg))
	}

	return outs, nil
}

// ScanSA scans for ServiceAccount refs.
func (p *Pod) ScanSA(ctx context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := p.getFactory().List(p.GVR(), ns, wait, labels.Everything())
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
		if serviceAccountMatches(pod.Spec.ServiceAccountName, n) {
			refs = append(refs, Ref{
				GVR: p.GVR(),
				FQN: client.FQN(pod.Namespace, pod.Name),
			})
		}
	}

	return refs, nil
}

// Scan scans for cluster resource refs.
func (p *Pod) Scan(ctx context.Context, gvr client.GVR, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := p.getFactory().List(p.GVR(), ns, wait, labels.Everything())
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
		case CmGVR:
			if !hasConfigMap(&pod.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: p.GVR(),
				FQN: client.FQN(pod.Namespace, pod.Name),
			})
		case SecGVR:
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
		case PvcGVR:
			if !hasPVC(&pod.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: p.GVR(),
				FQN: client.FQN(pod.Namespace, pod.Name),
			})
		case PcGVR:
			if !hasPC(&pod.Spec, n) {
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

func tailLogs(ctx context.Context, logger Logger, opts *LogOptions) LogChan {
	var (
		out = make(LogChan, 2)
		wg  sync.WaitGroup
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		podOpts := opts.ToPodLogOptions()
		var stream io.ReadCloser
		for r := 0; r < logRetryCount; r++ {
			var e error
			req, err := logger.Logs(opts.Path, podOpts)
			if err == nil {
				// This call will block if nothing is in the stream!!
				if stream, err = req.Stream(ctx); err == nil {
					wg.Add(1)
					go readLogs(ctx, &wg, stream, out, opts)
					return
				}
				e = fmt.Errorf("stream logs failed %w for %s", err, opts.Info())
				log.Error().Err(e).Msg("logs-stream")
			} else {
				e = fmt.Errorf("stream logs failed %w for %s", err, opts.Info())
				log.Error().Err(e).Msg("log-request")
			}

			select {
			case <-ctx.Done():
				return
			default:
				if e != nil {
					out <- opts.ToErrLogItem(e)
				}
				time.Sleep(logRetryWait)
			}
		}
	}()
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func readLogs(ctx context.Context, wg *sync.WaitGroup, stream io.ReadCloser, out chan<- *LogItem, opts *LogOptions) {
	defer func() {
		if err := stream.Close(); err != nil {
			log.Error().Err(err).Msgf("Fail to close stream %s", opts.Info())
		}
		wg.Done()
	}()

	log.Debug().Msgf(">>> LOG-READER PROCESSING %#v", opts)
	r := bufio.NewReader(stream)
	for {
		var item *LogItem
		if bytes, err := r.ReadBytes('\n'); err == nil {
			item = opts.ToLogItem(tview.EscapeBytes(bytes))
		} else {
			if errors.Is(err, io.EOF) {
				e := fmt.Errorf("stream closed %w for %s", err, opts.Info())
				item = opts.ToErrLogItem(e)
				log.Warn().Err(e).Msg("log-reader EOF")
			} else {
				e := fmt.Errorf("stream canceled %w for %s", err, opts.Info())
				item = opts.ToErrLogItem(e)
				log.Warn().Err(e).Msg("log-reader canceled")
			}
		}
		select {
		case <-ctx.Done():
			return
		case out <- item:
			if item.IsError {
				return
			}
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
	auth, err := p.Client().CanI(ns, "v1/pod", n, client.PatchAccess)
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
		return fmt.Errorf("unable to set image. This pod is managed by %s. Please set the image on the controller", manager)
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

func (p *Pod) Sanitize(ctx context.Context, ns string) (int, error) {
	oo, err := p.Resource.List(ctx, ns)
	if err != nil {
		return 0, err
	}

	var count int
	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			continue
		}
		var pod v1.Pod
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &pod)
		if err != nil {
			continue
		}
		log.Debug().Msgf("Pod status: %q", render.PodStatus(&pod))
		switch render.PodStatus(&pod) {
		case render.PhaseCompleted:
			fallthrough
		case render.PhasePending:
			fallthrough
		case render.PhaseCrashLoop:
			fallthrough
		case render.PhaseError:
			fallthrough
		case render.PhaseImagePullBackOff:
			fallthrough
		case render.PhaseContainerStatusUnknown:
			fallthrough
		case render.PhaseEvicted:
			fallthrough
		case render.PhaseOOMKilled:
			// !!BOZO!! Might need to bump timeout otherwise rev limit if too many??
			log.Debug().Msgf("Sanitizing %s:%s", pod.Namespace, pod.Name)
			fqn := client.FQN(pod.Namespace, pod.Name)
			if err := p.Delete(ctx, fqn, nil, 0); err != nil {
				log.Debug().Msgf("Aborted! Sanitizer deleted %d pods", count)
				return count, err
			}
			count++
		}
	}
	log.Debug().Msgf("Sanitizer deleted %d pods", count)

	return count, nil
}
