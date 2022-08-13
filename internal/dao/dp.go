package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
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
)

// Deployment represents a deployment K8s resource.
type Deployment struct {
	Resource
}

// IsHappy check for happy deployments.
func (d *Deployment) IsHappy(dp appsv1.Deployment) bool {
	return dp.Status.Replicas == dp.Status.AvailableReplicas
}

// Scale a Deployment.
func (d *Deployment) Scale(ctx context.Context, path string, replicas int32) error {
	ns, n := client.Namespaced(path)
	auth, err := d.Client().CanI(ns, "apps/v1/deployments:scale", []string{client.GetVerb, client.UpdateVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to scale a deployment")
	}

	dial, err := d.Client().Dial()
	if err != nil {
		return err
	}
	scale, err := dial.AppsV1().Deployments(ns).GetScale(ctx, n, metav1.GetOptions{})
	if err != nil {
		return err
	}
	scale.Spec.Replicas = replicas
	_, err = dial.AppsV1().Deployments(ns).UpdateScale(ctx, n, scale, metav1.UpdateOptions{})

	return err
}

// Restart a Deployment rollout.
func (d *Deployment) Restart(ctx context.Context, path string) error {
	o, err := d.GetFactory().Get("apps/v1/deployments", path, true, labels.Everything())
	if err != nil {
		return err
	}
	var dp appsv1.Deployment
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &dp)
	if err != nil {
		return err
	}

	auth, err := d.Client().CanI(dp.Namespace, "apps/v1/deployments", []string{client.PatchVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to restart a deployment")
	}

	dial, err := d.Client().Dial()
	if err != nil {
		return err
	}

	before, err := runtime.Encode(scheme.Codecs.LegacyCodec(appsv1.SchemeGroupVersion), &dp)
	if err != nil {
		return err
	}

	after, err := polymorphichelpers.ObjectRestarterFn(&dp)
	if err != nil {
		return err
	}
	diff, err := strategicpatch.CreateTwoWayMergePatch(before, after, dp)
	if err != nil {
		return err
	}
	_, err = dial.AppsV1().Deployments(dp.Namespace).Patch(
		ctx,
		dp.Name,
		types.StrategicMergePatchType,
		diff,
		metav1.PatchOptions{},
	)

	return err
}

// TailLogs tail logs for all pods represented by this Deployment.
func (d *Deployment) TailLogs(ctx context.Context, opts *LogOptions) ([]LogChan, error) {
	dp, err := d.Load(d.Factory, opts.Path)
	if err != nil {
		return nil, err
	}
	if dp.Spec.Selector == nil || len(dp.Spec.Selector.MatchLabels) == 0 {
		return nil, fmt.Errorf("No valid selector found on Deployment %s", opts.Path)
	}

	return podLogs(ctx, dp.Spec.Selector.MatchLabels, opts)
}

// Pod returns a pod victim by name.
func (d *Deployment) Pod(fqn string) (string, error) {
	dp, err := d.Load(d.Factory, fqn)
	if err != nil {
		return "", err
	}

	return podFromSelector(d.Factory, dp.Namespace, dp.Spec.Selector.MatchLabels)
}

// Load returns a deployment instance.
func (*Deployment) Load(f Factory, fqn string) (*appsv1.Deployment, error) {
	o, err := f.Get("apps/v1/deployments", fqn, true, labels.Everything())
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
func (d *Deployment) ScanSA(ctx context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := d.GetFactory().List(d.GVR(), ns, wait, labels.Everything())
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
		if dp.Spec.Template.Spec.ServiceAccountName == n {
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(dp.Namespace, dp.Name),
			})
		}
	}

	return refs, nil
}

// Scan scans for resource references.
func (d *Deployment) Scan(ctx context.Context, gvr, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := d.GetFactory().List(d.GVR(), ns, wait, labels.Everything())
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
		case "v1/configmaps":
			if !hasConfigMap(&dp.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(dp.Namespace, dp.Name),
			})
		case "v1/secrets":
			found, err := hasSecret(d.Factory, &dp.Spec.Template.Spec, dp.Namespace, n, wait)
			if err != nil {
				log.Warn().Err(err).Msgf("scanning secret %q", fqn)
				continue
			}
			if !found {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(dp.Namespace, dp.Name),
			})
		case "v1/persistentvolumeclaims":
			if !hasPVC(&dp.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(dp.Namespace, dp.Name),
			})
		case "scheduling.k8s.io/v1/priorityclasses":
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
	dp, err := d.Load(d.Factory, path)
	if err != nil {
		return nil, err
	}
	podSpec := dp.Spec.Template.Spec
	return &podSpec, nil
}

// SetImages sets container images.
func (d *Deployment) SetImages(ctx context.Context, path string, imageSpecs ImageSpecs) error {
	ns, n := client.Namespaced(path)
	auth, err := d.Client().CanI(ns, "apps/v1/deployments", []string{client.PatchVerb})
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

func hasPVC(spec *v1.PodSpec, name string) bool {
	for _, v := range spec.Volumes {
		if v.PersistentVolumeClaim != nil && v.PersistentVolumeClaim.ClaimName == name {
			return true
		}
	}
	return false
}

func hasPC(spec *v1.PodSpec, name string) bool {
	return spec.PriorityClassName == name
}

func hasConfigMap(spec *v1.PodSpec, name string) bool {
	for _, c := range spec.InitContainers {
		if containerHasConfigMap(c, name) {
			return true
		}
	}
	for _, c := range spec.Containers {
		if containerHasConfigMap(c, name) {
			return true
		}
	}

	for _, v := range spec.Volumes {
		if cm := v.VolumeSource.ConfigMap; cm != nil {
			if cm.LocalObjectReference.Name == name {
				return true
			}
		}
	}
	return false
}

// BOZO !! Need to deal with ephemeral containers.
func hasSecret(f Factory, spec *v1.PodSpec, ns, name string, wait bool) (bool, error) {
	for _, c := range spec.InitContainers {
		if containerHasSecret(c, name) {
			return true, nil
		}
	}
	for _, c := range spec.Containers {
		if containerHasSecret(c, name) {
			return true, nil
		}
	}

	saName := spec.ServiceAccountName
	if saName != "" {
		o, err := f.Get("v1/serviceaccounts", client.FQN(ns, saName), wait, labels.Everything())
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

	for _, v := range spec.Volumes {
		if sec := v.VolumeSource.Secret; sec != nil {
			if sec.SecretName == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func containerHasSecret(c v1.Container, name string) bool {
	for _, e := range c.EnvFrom {
		if e.SecretRef != nil && e.SecretRef.Name == name {
			return true
		}
	}
	for _, e := range c.Env {
		if e.ValueFrom == nil || e.ValueFrom.SecretKeyRef == nil {
			continue
		}
		if e.ValueFrom.SecretKeyRef.Name == name {
			return true
		}
	}

	return false
}

func containerHasConfigMap(c v1.Container, name string) bool {
	for _, e := range c.EnvFrom {
		if e.ConfigMapRef != nil && e.ConfigMapRef.Name == name {
			return true
		}
	}
	for _, e := range c.Env {
		if e.ValueFrom == nil || e.ValueFrom.ConfigMapKeyRef == nil {
			continue
		}
		if e.ValueFrom.ConfigMapKeyRef.Name == name {
			return true
		}
	}

	return false
}
