// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
)

var (
	_ Accessor        = (*Deployment)(nil)
	_ Nuker           = (*Deployment)(nil)
	_ Loggable        = (*Deployment)(nil)
	_ Restartable     = (*Deployment)(nil)
	_ Scalable        = (*Deployment)(nil)
	_ Controller      = (*Deployment)(nil)
	_ ContainsPodSpec = (*Deployment)(nil)
	_ ImageLister     = (*Deployment)(nil)
)

// Deployment represents a deployment K8s resource.
type Deployment struct {
	Resource
}

// ListImages lists container images.
func (d *Deployment) ListImages(_ context.Context, fqn string) ([]string, error) {
	dp, err := d.GetInstance(fqn)
	if err != nil {
		return nil, err
	}

	return render.ExtractImages(&dp.Spec.Template.Spec), nil
}

// Scale a Deployment.
func (d *Deployment) Scale(ctx context.Context, path string, replicas int32) error {
	return scaleRes(ctx, d.getFactory(), client.DpGVR, path, replicas)
}

// Restart a Deployment rollout.
func (d *Deployment) Restart(ctx context.Context, path string, opts *metav1.PatchOptions) error {
	return restartRes[*appsv1.Deployment](ctx, d.getFactory(), client.DpGVR, path, opts)
}

// TailLogs tail logs for all pods represented by this Deployment.
func (d *Deployment) TailLogs(ctx context.Context, opts *LogOptions) ([]LogChan, error) {
	dp, err := d.GetInstance(opts.Path)
	if err != nil {
		return nil, err
	}
	if dp.Spec.Selector == nil || len(dp.Spec.Selector.MatchLabels) == 0 {
		return nil, fmt.Errorf("no valid selector found on deployment: %s", opts.Path)
	}

	return podLogs(ctx, dp.Spec.Selector.MatchLabels, opts)
}

// Pod returns a pod victim by name.
func (d *Deployment) Pod(fqn string) (string, error) {
	dp, err := d.GetInstance(fqn)
	if err != nil {
		return "", err
	}

	return podFromSelector(d.Factory, dp.Namespace, dp.Spec.Selector.MatchLabels)
}

// GetInstance fetch a matching deployment.
func (d *Deployment) GetInstance(fqn string) (*appsv1.Deployment, error) {
	o, err := d.Factory.Get(d.gvr, fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var dp appsv1.Deployment
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &dp)
	if err != nil {
		return nil, errors.New("expecting Deployment resource")
	}

	return &dp, nil
}

// ScanSA scans for serviceaccount refs.
func (d *Deployment) ScanSA(_ context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := d.getFactory().List(d.gvr, ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var dp appsv1.Deployment
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &dp)
		if err != nil {
			return nil, errors.New("expecting Deployment resource")
		}
		if serviceAccountMatches(dp.Spec.Template.Spec.ServiceAccountName, n) {
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(dp.Namespace, dp.Name),
			})
		}
	}

	return refs, nil
}

// Scan scans for resource references.
func (d *Deployment) Scan(_ context.Context, gvr *client.GVR, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := d.getFactory().List(d.gvr, ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var dp appsv1.Deployment
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &dp)
		if err != nil {
			return nil, errors.New("expecting Deployment resource")
		}
		switch gvr {
		case client.CmGVR:
			if !hasConfigMap(&dp.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(dp.Namespace, dp.Name),
			})
		case client.SecGVR:
			found, err := hasSecret(d.Factory, &dp.Spec.Template.Spec, dp.Namespace, n, wait)
			if err != nil {
				slog.Warn("Fail to locate secret",
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
				FQN: client.FQN(dp.Namespace, dp.Name),
			})
		case client.PvcGVR:
			if !hasPVC(&dp.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(dp.Namespace, dp.Name),
			})
		case client.PcGVR:
			if !hasPC(&dp.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(dp.Namespace, dp.Name),
			})
		}
	}

	return refs, nil
}

// GetPodSpec returns a pod spec given a resource.
func (d *Deployment) GetPodSpec(path string) (*v1.PodSpec, error) {
	dp, err := d.GetInstance(path)
	if err != nil {
		return nil, err
	}
	podSpec := dp.Spec.Template.Spec
	return &podSpec, nil
}

// SetImages sets container images.
func (d *Deployment) SetImages(ctx context.Context, path string, imageSpecs ImageSpecs) error {
	ns, n := client.Namespaced(path)
	auth, err := d.Client().CanI(ns, d.gvr, n, client.PatchAccess)
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to patch a deployment")
	}
	jsonPatch, err := GetTemplateJsonPatch(imageSpecs)
	if err != nil {
		return err
	}
	dial, err := d.Client().Dial()
	if err != nil {
		return err
	}
	_, err = dial.AppsV1().Deployments(ns).Patch(
		ctx,
		n,
		types.StrategicMergePatchType,
		jsonPatch,
		metav1.PatchOptions{},
	)
	return err
}

// Helpers...

func hasPVC(spec *v1.PodSpec, name string) bool {
	for i := range spec.Volumes {
		if spec.Volumes[i].PersistentVolumeClaim != nil && spec.Volumes[i].PersistentVolumeClaim.ClaimName == name {
			return true
		}
	}
	return false
}

func hasPC(spec *v1.PodSpec, name string) bool {
	return spec.PriorityClassName == name
}

func hasConfigMap(spec *v1.PodSpec, name string) bool {
	for i := range spec.InitContainers {
		if containerHasConfigMap(spec.InitContainers[i].EnvFrom, spec.InitContainers[i].Env, name) {
			return true
		}
	}
	for i := range spec.Containers {
		if containerHasConfigMap(spec.Containers[i].EnvFrom, spec.Containers[i].Env, name) {
			return true
		}
	}
	for i := range spec.EphemeralContainers {
		if containerHasConfigMap(spec.EphemeralContainers[i].EnvFrom, spec.EphemeralContainers[i].Env, name) {
			return true
		}
	}

	for i := range spec.Volumes {
		if cm := spec.Volumes[i].ConfigMap; cm != nil {
			if cm.Name == name {
				return true
			}
		}
	}
	return false
}

func hasSecret(f Factory, spec *v1.PodSpec, ns, name string, wait bool) (bool, error) {
	for i := range spec.InitContainers {
		if containerHasSecret(spec.InitContainers[i].EnvFrom, spec.InitContainers[i].Env, name) {
			return true, nil
		}
	}

	for i := range spec.Containers {
		if containerHasSecret(spec.Containers[i].EnvFrom, spec.Containers[i].Env, name) {
			return true, nil
		}
	}

	for i := range spec.EphemeralContainers {
		if containerHasSecret(spec.EphemeralContainers[i].EnvFrom, spec.EphemeralContainers[i].Env, name) {
			return true, nil
		}
	}

	for _, s := range spec.ImagePullSecrets {
		if s.Name == name {
			return true, nil
		}
	}

	if saName := spec.ServiceAccountName; saName != "" {
		o, err := f.Get(client.SaGVR, client.FQN(ns, saName), wait, labels.Everything())
		if err != nil {
			return false, err
		}

		var sa v1.ServiceAccount
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sa)
		if err != nil {
			return false, errors.New("expecting ServiceAccount resource")
		}

		for _, ref := range sa.Secrets {
			if ref.Namespace == ns && ref.Name == name {
				return true, nil
			}
		}
	}

	for i := range spec.Volumes {
		if sec := spec.Volumes[i].Secret; sec != nil {
			if sec.SecretName == name {
				return true, nil
			}
		}
	}

	return false, nil
}

func containerHasSecret(envFrom []v1.EnvFromSource, env []v1.EnvVar, name string) bool {
	for _, e := range envFrom {
		if e.SecretRef != nil && e.SecretRef.Name == name {
			return true
		}
	}
	for _, e := range env {
		if e.ValueFrom == nil || e.ValueFrom.SecretKeyRef == nil {
			continue
		}
		if e.ValueFrom.SecretKeyRef.Name == name {
			return true
		}
	}

	return false
}

func containerHasConfigMap(envFrom []v1.EnvFromSource, env []v1.EnvVar, name string) bool {
	for _, e := range envFrom {
		if e.ConfigMapRef != nil && e.ConfigMapRef.Name == name {
			return true
		}
	}
	for _, e := range env {
		if e.ValueFrom == nil || e.ValueFrom.ConfigMapKeyRef == nil {
			continue
		}
		if e.ValueFrom.ConfigMapKeyRef.Name == name {
			return true
		}
	}

	return false
}

func scaleRes(ctx context.Context, f Factory, gvr *client.GVR, path string, replicas int32) error {
	ns, n := client.Namespaced(path)
	auth, err := f.Client().CanI(ns, client.NewGVR(gvr.String()+":scale"), n, []string{client.GetVerb, client.UpdateVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to scale: %s", gvr)
	}

	dial, err := f.Client().Dial()
	if err != nil {
		return err
	}

	switch gvr {
	case client.DpGVR:
		scale, e := dial.AppsV1().Deployments(ns).GetScale(ctx, n, metav1.GetOptions{})
		if e != nil {
			return e
		}
		scale.Spec.Replicas = replicas
		_, e = dial.AppsV1().Deployments(ns).UpdateScale(ctx, n, scale, metav1.UpdateOptions{})
		return e
	case client.StsGVR:
		scale, e := dial.AppsV1().StatefulSets(ns).GetScale(ctx, n, metav1.GetOptions{})
		if e != nil {
			return e
		}
		scale.Spec.Replicas = replicas
		_, e = dial.AppsV1().StatefulSets(ns).UpdateScale(ctx, n, scale, metav1.UpdateOptions{})
		return e
	default:
		return fmt.Errorf("unsupported resource for scaling: %s", gvr)
	}
}

func restartRes[T runtime.Object](ctx context.Context, f Factory, gvr *client.GVR, path string, opts *metav1.PatchOptions) error {
	o, err := f.Get(gvr, path, true, labels.Everything())
	if err != nil {
		return err
	}
	var r = new(T)
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, r)
	if err != nil {
		return err
	}

	ns, n := client.Namespaced(path)
	auth, err := f.Client().CanI(ns, gvr, n, client.PatchAccess)
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to restart %q", gvr)
	}

	dial, err := f.Client().Dial()
	if err != nil {
		return err
	}

	before, err := runtime.Encode(scheme.Codecs.LegacyCodec(appsv1.SchemeGroupVersion), *r)
	if err != nil {
		return err
	}
	after, err := polymorphichelpers.ObjectRestarterFn(*r)
	if err != nil {
		return err
	}
	diff, err := strategicpatch.CreateTwoWayMergePatch(before, after, *r)
	if err != nil {
		return err
	}

	switch gvr {
	case client.DpGVR:
		_, err = dial.AppsV1().Deployments(ns).Patch(
			ctx,
			n,
			types.StrategicMergePatchType,
			diff,
			*opts,
		)

	case client.DsGVR:
		_, err = dial.AppsV1().DaemonSets(ns).Patch(
			ctx,
			n,
			types.StrategicMergePatchType,
			diff,
			*opts,
		)

	case client.StsGVR:
		_, err = dial.AppsV1().StatefulSets(ns).Patch(
			ctx,
			n,
			types.StrategicMergePatchType,
			diff,
			*opts,
		)
	}

	return err
}
