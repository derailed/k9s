// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/watch"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var (
	_ Accessor        = (*DaemonSet)(nil)
	_ Nuker           = (*DaemonSet)(nil)
	_ Loggable        = (*DaemonSet)(nil)
	_ Restartable     = (*DaemonSet)(nil)
	_ Controller      = (*DaemonSet)(nil)
	_ ContainsPodSpec = (*DaemonSet)(nil)
	_ ImageLister     = (*DaemonSet)(nil)
)

// DaemonSet represents a K8s daemonset.
type DaemonSet struct {
	Resource
}

// ListImages lists container images.
func (d *DaemonSet) ListImages(_ context.Context, fqn string) ([]string, error) {
	ds, err := d.GetInstance(fqn)
	if err != nil {
		return nil, err
	}

	return render.ExtractImages(&ds.Spec.Template.Spec), nil
}

// Restart a DaemonSet rollout.
func (d *DaemonSet) Restart(ctx context.Context, path string) error {
	return restartRes[*appsv1.DaemonSet](ctx, d.getFactory(), client.DsGVR, path)
}

// TailLogs tail logs for all pods represented by this DaemonSet.
func (d *DaemonSet) TailLogs(ctx context.Context, opts *LogOptions) ([]LogChan, error) {
	ds, err := d.GetInstance(opts.Path)
	if err != nil {
		return nil, err
	}

	if ds.Spec.Selector == nil || len(ds.Spec.Selector.MatchLabels) == 0 {
		return nil, fmt.Errorf("no valid selector found on daemonset %q", opts.Path)
	}

	return podLogs(ctx, ds.Spec.Selector.MatchLabels, opts)
}

func podLogs(ctx context.Context, sel map[string]string, opts *LogOptions) ([]LogChan, error) {
	f, ok := ctx.Value(internal.KeyFactory).(*watch.Factory)
	if !ok {
		return nil, errors.New("expecting a context factory")
	}
	ls, err := metav1.ParseToLabelSelector(toSelector(sel))
	if err != nil {
		return nil, err
	}
	lsel, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		return nil, err
	}

	ns, _ := client.Namespaced(opts.Path)
	oo, err := f.List(client.PodGVR, ns, true, lsel)
	if err != nil {
		return nil, err
	}
	opts.MultiPods = true

	var po Pod
	po.Init(f, client.PodGVR)

	outs := make([]LogChan, 0, len(oo))
	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("expected unstructured got %t", o)
		}
		opts = opts.Clone()
		opts.Path = client.FQN(u.GetNamespace(), u.GetName())
		cc, err := po.TailLogs(ctx, opts)
		if err != nil {
			return nil, err
		}
		outs = append(outs, cc...)
	}

	return outs, nil
}

// Pod returns a pod victim by name.
func (d *DaemonSet) Pod(fqn string) (string, error) {
	ds, err := d.GetInstance(fqn)
	if err != nil {
		return "", err
	}

	return podFromSelector(d.Factory, ds.Namespace, ds.Spec.Selector.MatchLabels)
}

// GetInstance returns a daemonset instance.
func (d *DaemonSet) GetInstance(fqn string) (*appsv1.DaemonSet, error) {
	o, err := d.getFactory().Get(d.gvr, fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var ds appsv1.DaemonSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
	if err != nil {
		return nil, errors.New("expecting DaemonSet resource")
	}

	return &ds, nil
}

// ScanSA scans for serviceaccount refs.
func (d *DaemonSet) ScanSA(_ context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := d.getFactory().List(d.gvr, ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var ds appsv1.DaemonSet
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
		if err != nil {
			return nil, errors.New("expecting DaemonSet resource")
		}
		if serviceAccountMatches(ds.Spec.Template.Spec.ServiceAccountName, n) {
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(ds.Namespace, ds.Name),
			})
		}
	}

	return refs, nil
}

// Scan scans for cluster refs.
func (d *DaemonSet) Scan(_ context.Context, gvr *client.GVR, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := d.getFactory().List(d.gvr, ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var ds appsv1.DaemonSet
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
		if err != nil {
			return nil, errors.New("expecting StatefulSet resource")
		}
		switch gvr {
		case client.CmGVR:
			if !hasConfigMap(&ds.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(ds.Namespace, ds.Name),
			})
		case client.SecGVR:
			found, err := hasSecret(d.Factory, &ds.Spec.Template.Spec, ds.Namespace, n, wait)
			if err != nil {
				slog.Warn("Unable to locate secret",
					slogs.FQN, fqn,
					slogs.Error, err,
				)
				continue
			}
			if !found {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(ds.Namespace, ds.Name),
			})
		case client.PvcGVR:
			if !hasPVC(&ds.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(ds.Namespace, ds.Name),
			})
		case client.PcGVR:
			if !hasPC(&ds.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(ds.Namespace, ds.Name),
			})
		}
	}

	return refs, nil
}

// GetPodSpec returns a pod spec given a resource.
func (d *DaemonSet) GetPodSpec(path string) (*v1.PodSpec, error) {
	ds, err := d.GetInstance(path)
	if err != nil {
		return nil, err
	}
	podSpec := ds.Spec.Template.Spec
	return &podSpec, nil
}

// SetImages sets container images.
func (d *DaemonSet) SetImages(ctx context.Context, path string, imageSpecs ImageSpecs) error {
	ns, n := client.Namespaced(path)
	auth, err := d.Client().CanI(ns, d.gvr, n, client.PatchAccess)
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to patch a daemonset")
	}
	jsonPatch, err := GetTemplateJsonPatch(imageSpecs)
	if err != nil {
		return err
	}
	dial, err := d.Client().Dial()
	if err != nil {
		return err
	}
	_, err = dial.AppsV1().DaemonSets(ns).Patch(
		ctx,
		n,
		types.StrategicMergePatchType,
		jsonPatch,
		metav1.PatchOptions{},
	)
	return err
}

// ----------------------------------------------------------------------------
// Helpers...

func toSelector(m map[string]string) string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		s = append(s, k+"="+v)
	}

	return strings.Join(s, ",")
}
