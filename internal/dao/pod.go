// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tview"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
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

type streamResult int

const (
	logRetryCount                  = 20
	logBackoffInitial              = 500 * time.Millisecond
	logBackoffMax                  = 30 * time.Second
	logChannelBuffer               = 50   // Buffer size for log channel to reduce drops
	streamEOF         streamResult = iota // legit container log close (no retry)
	streamError                           // retryable error (network, auth, etc.)
	streamCanceled                        // context canceled
)

// Pod represents a pod resource.
type Pod struct {
	Resource
}

// shouldStopRetrying checks if we should stop retrying log streaming based on pod status.
func (p *Pod) shouldStopRetrying(path string) bool {
	pod, err := p.GetInstance(path)
	if err != nil {
		return true
	}

	if pod.DeletionTimestamp != nil {
		return true
	}

	switch pod.Status.Phase {
	case v1.PodSucceeded, v1.PodFailed:
		return true
	}

	return false
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
func (p *Pod) ListImages(_ context.Context, path string) ([]string, error) {
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

		spec, ok := u.Object["spec"].(map[string]any)
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
	auth, err := p.Client().CanI(ns, client.NewGVR(client.PodGVR.String()+":log"), n, client.GetAccess)
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
	for i := range pod.Spec.Containers {
		cc = append(cc, pod.Spec.Containers[i].Name)
	}

	if includeInit {
		for i := range pod.Spec.InitContainers {
			cc = append(cc, pod.Spec.InitContainers[i].Name)
		}
	}

	return cc, nil
}

// Pod returns a pod victim by name.
func (*Pod) Pod(fqn string) (string, error) {
	return fqn, nil
}

// GetInstance returns a pod instance.
func (p *Pod) GetInstance(fqn string) (*v1.Pod, error) {
	o, err := p.getFactory().Get(p.gvr, fqn, true, labels.Everything())
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
	o, err := fac.Get(p.gvr, opts.Path, true, labels.Everything())
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
	if co, ok := GetDefaultContainer(&po.ObjectMeta, &po.Spec); ok && !opts.AllContainers {
		opts.DefaultContainer = co
		return append(outs, tailLogs(ctx, p, opts)), nil
	}
	if opts.HasContainer() && !opts.AllContainers {
		return append(outs, tailLogs(ctx, p, opts)), nil
	}
	for i := range po.Spec.InitContainers {
		cfg := opts.Clone()
		cfg.Container = po.Spec.InitContainers[i].Name
		outs = append(outs, tailLogs(ctx, p, cfg))
	}
	for i := range po.Spec.Containers {
		cfg := opts.Clone()
		cfg.Container = po.Spec.Containers[i].Name
		outs = append(outs, tailLogs(ctx, p, cfg))
	}
	for i := range po.Spec.EphemeralContainers {
		cfg := opts.Clone()
		cfg.Container = po.Spec.EphemeralContainers[i].Name
		outs = append(outs, tailLogs(ctx, p, cfg))
	}

	return outs, nil
}

// ScanSA scans for ServiceAccount refs.
func (p *Pod) ScanSA(_ context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := p.getFactory().List(p.gvr, ns, wait, labels.Everything())
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
		if len(pod.OwnerReferences) > 0 {
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
func (p *Pod) Scan(_ context.Context, gvr *client.GVR, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := p.getFactory().List(p.gvr, ns, wait, labels.Everything())
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
		if len(pod.OwnerReferences) > 0 {
			continue
		}
		switch gvr {
		case client.CmGVR:
			if !hasConfigMap(&pod.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: p.GVR(),
				FQN: client.FQN(pod.Namespace, pod.Name),
			})
		case client.SecGVR:
			found, err := hasSecret(p.Factory, &pod.Spec, pod.Namespace, n, wait)
			if err != nil {
				slog.Warn("Locate secret failed",
					slogs.FQN, fqn,
					slogs.Error, err,
				)
				continue
			}
			if !found {
				continue
			}
			refs = append(refs, Ref{
				GVR: p.GVR(),
				FQN: client.FQN(pod.Namespace, pod.Name),
			})
		case client.PvcGVR:
			if !hasPVC(&pod.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: p.GVR(),
				FQN: client.FQN(pod.Namespace, pod.Name),
			})
		case client.PcGVR:
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
	out := make(LogChan, logChannelBuffer)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		podOpts := opts.ToPodLogOptions()

		// Setup exponential backoff following project pattern
		bf := backoff.NewExponentialBackOff()
		bf.InitialInterval = logBackoffInitial
		bf.MaxElapsedTime = 0
		bf.MaxInterval = logBackoffMax / 2
		backoffCtx := backoff.WithContext(bf, ctx)
		delay := logBackoffInitial

		for range logRetryCount {
			// Check if we should stop retrying based on pod status
			if pod, ok := logger.(*Pod); ok && pod.shouldStopRetrying(opts.Path) {
				slog.Debug("Stopping log retry - pod is terminating or deleted",
					slogs.Container, opts.Info(),
				)
				return
			}

			req, err := logger.Logs(opts.Path, podOpts)
			if err != nil {
				slog.Error("Log request failed",
					slogs.Container, opts.Info(),
					slogs.Error, err,
				)
				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
					if delay = backoffCtx.NextBackOff(); delay == backoff.Stop {
						return
					}
				}
				continue
			}

			stream, e := req.Stream(ctx)
			if e != nil {
				slog.Error("Stream logs failed",
					slogs.Error, e,
					slogs.Container, opts.Info(),
				)
				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
					if delay = backoffCtx.NextBackOff(); delay == backoff.Stop {
						return
					}
				}
				continue
			}

			// Process logs until completion
			result := readLogs(ctx, stream, out, opts)
			switch result {
			case streamEOF:
				slog.Debug("Log stream ended cleanly",
					slogs.Container, opts.Info(),
				)
				return
			case streamError:
				// Check if we should stop retrying based on pod status
				if pod, ok := logger.(*Pod); ok && pod.shouldStopRetrying(opts.Path) {
					slog.Debug("Stopping log retry after stream error - pod is terminating or deleted",
						slogs.Container, opts.Info(),
					)
					return
				}
				slog.Debug("Log stream error, retrying",
					slogs.Container, opts.Info(),
				)
				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
					if delay = backoffCtx.NextBackOff(); delay == backoff.Stop {
						return
					}
				}
				continue
			case streamCanceled:
				return
			}

			// Reset backoff and delay on successful connection
			bf.Reset()
			delay = logBackoffInitial
		}

		// Out of retries
		out <- opts.ToErrLogItem(fmt.Errorf("failed to maintain log stream after %d retries", logRetryCount))
	}()

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func readLogs(ctx context.Context, stream io.ReadCloser, out chan<- *LogItem, opts *LogOptions) streamResult {
	defer func() {
		if err := stream.Close(); err != nil && !errors.Is(err, io.ErrClosedPipe) {
			slog.Error("Fail to close stream",
				slogs.Container, opts.Info(),
				slogs.Error, err,
			)
		}
	}()

	r := bufio.NewReader(stream)

	for {
		bytes, err := r.ReadBytes('\n')
		if err == nil {
			item := opts.ToLogItem(tview.EscapeBytes(bytes))
			select {
			case <-ctx.Done():
				return streamCanceled
			case out <- item:
			default:
				// Avoid deadlock if consumer is too slow
				slog.Warn("Dropping log line due to slow consumer",
					slogs.Container, opts.Info(),
				)
			}
			continue
		}

		if errors.Is(err, io.EOF) {
			if len(bytes) > 0 {
				// Emit trailing partial line before EOF
				out <- opts.ToLogItem(tview.EscapeBytes(bytes))
			}
			slog.Debug("Log reader reached EOF", slogs.Container, opts.Info())
			out <- opts.ToErrLogItem(fmt.Errorf("stream closed: %w for %s", err, opts.Info()))
			return streamEOF
		}

		// Non-EOF error
		slog.Debug("Log stream error, will retry connection",
			slogs.Container, opts.Info(),
			slogs.Error, fmt.Errorf("stream error: %w for %s", err, opts.Info()),
		)
		// Don't send stream errors to user - they will be retried
		// Only final retry exhaustion message is shown
		return streamError
	}
}

// MetaFQN returns a fully qualified resource name.
func MetaFQN(m *metav1.ObjectMeta) string {
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
	auth, err := p.Client().CanI(ns, p.gvr, n, client.PatchAccess)
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

func (p *Pod) isControlled(path string) (fqn string, ok bool, err error) {
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

var toastPhases = sets.New(
	render.PhaseCompleted,
	render.PhasePending,
	render.PhaseCrashLoop,
	render.PhaseError,
	render.PhaseImagePullBackOff,
	render.PhaseContainerStatusUnknown,
	render.PhaseEvicted,
	render.PhaseOOMKilled,
)

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

		if toastPhases.Has(render.PodStatus(&pod)) {
			// !!BOZO!! Might need to bump timeout otherwise rev limit if too many??
			fqn := client.FQN(pod.Namespace, pod.Name)
			slog.Debug("Sanitizing resource", slogs.FQN, fqn)
			if err := p.Delete(ctx, fqn, nil, 0); err != nil {
				slog.Debug("Aborted! Sanitizer delete failed",
					slogs.FQN, fqn,
					slogs.Count, count,
					slogs.Error, err,
				)
				return count, err
			}
			count++
		}
	}
	slog.Debug("Sanitizer deleted pods", slogs.Count, count)

	return count, nil
}
