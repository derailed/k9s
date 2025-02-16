// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
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
	_ Accessor        = (*StatefulSet)(nil)
	_ Nuker           = (*StatefulSet)(nil)
	_ Loggable        = (*StatefulSet)(nil)
	_ Restartable     = (*StatefulSet)(nil)
	_ Scalable        = (*StatefulSet)(nil)
	_ Controller      = (*StatefulSet)(nil)
	_ ContainsPodSpec = (*StatefulSet)(nil)
	_ ImageLister     = (*StatefulSet)(nil)
)

// StatefulSet represents a K8s sts.
type StatefulSet struct {
	Resource
}

// ListImages lists container images.
func (s *StatefulSet) ListImages(ctx context.Context, fqn string) ([]string, error) {
	sts, err := s.GetInstance(s.Factory, fqn)
	if err != nil {
		return nil, err
	}

	return render.ExtractImages(&sts.Spec.Template.Spec), nil
}

// Scale a StatefulSet.
func (s *StatefulSet) Scale(ctx context.Context, path string, replicas int32) error {
	ns, n := client.Namespaced(path)
	auth, err := s.Client().CanI(ns, "apps/v1/statefulsets:scale", n, []string{client.GetVerb, client.UpdateVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to scale statefulsets")
	}

	dial, err := s.Client().Dial()
	if err != nil {
		return err
	}
	scale, err := dial.AppsV1().StatefulSets(ns).GetScale(ctx, n, metav1.GetOptions{})
	if err != nil {
		return err
	}
	scale.Spec.Replicas = replicas
	_, err = dial.AppsV1().StatefulSets(ns).UpdateScale(ctx, n, scale, metav1.UpdateOptions{})

	return err
}

// Restart a StatefulSet rollout.
func (s *StatefulSet) Restart(ctx context.Context, path string) error {
	sts, err := s.GetInstance(s.Factory, path)
	if err != nil {
		return err
	}

	ns, n := client.Namespaced(path)
	pp, err := podsFromSelector(s.Factory, ns, sts.Spec.Selector.MatchLabels)
	if err != nil {
		return err
	}
	for _, p := range pp {
		s.Forwarders().Kill(client.FQN(p.Namespace, p.Name))
	}

	auth, err := s.Client().CanI(sts.Namespace, "apps/v1/statefulsets", n, client.PatchAccess)
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to restart a statefulset")
	}

	dial, err := s.Client().Dial()
	if err != nil {
		return err
	}

	before, err := runtime.Encode(scheme.Codecs.LegacyCodec(appsv1.SchemeGroupVersion), sts)
	if err != nil {
		return err
	}

	after, err := polymorphichelpers.ObjectRestarterFn(sts)
	if err != nil {
		return err
	}
	diff, err := strategicpatch.CreateTwoWayMergePatch(before, after, sts)
	if err != nil {
		return err
	}
	_, err = dial.AppsV1().StatefulSets(sts.Namespace).Patch(
		ctx,
		sts.Name,
		types.StrategicMergePatchType,
		diff,
		metav1.PatchOptions{},
	)

	return err

}

// GetInstance returns a statefulset instance.
func (*StatefulSet) GetInstance(f Factory, fqn string) (*appsv1.StatefulSet, error) {
	o, err := f.Get("apps/v1/statefulsets", fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var sts appsv1.StatefulSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sts)
	if err != nil {
		return nil, errors.New("expecting Statefulset resource")
	}

	return &sts, nil
}

// TailLogs tail logs for all pods represented by this StatefulSet.
func (s *StatefulSet) TailLogs(ctx context.Context, opts *LogOptions) ([]LogChan, error) {
	sts, err := s.getStatefulSet(opts.Path)
	if err != nil {
		return nil, errors.New("expecting StatefulSet resource")
	}
	if sts.Spec.Selector == nil || len(sts.Spec.Selector.MatchLabels) == 0 {
		return nil, fmt.Errorf("no valid selector found on statefulset: %s", opts.Path)
	}

	return podLogs(ctx, sts.Spec.Selector.MatchLabels, opts)
}

// Pod returns a pod victim by name.
func (s *StatefulSet) Pod(fqn string) (string, error) {
	sts, err := s.getStatefulSet(fqn)
	if err != nil {
		return "", err
	}

	return podFromSelector(s.Factory, sts.Namespace, sts.Spec.Selector.MatchLabels)
}

func (s *StatefulSet) getStatefulSet(fqn string) (*appsv1.StatefulSet, error) {
	o, err := s.getFactory().Get(s.gvrStr(), fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var sts appsv1.StatefulSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sts)
	if err != nil {
		return nil, errors.New("expecting Service resource")
	}

	return &sts, nil
}

// ScanSA scans for serviceaccount refs.
func (s *StatefulSet) ScanSA(ctx context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := s.getFactory().List(s.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var sts appsv1.StatefulSet
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sts)
		if err != nil {
			return nil, errors.New("expecting StatefulSet resource")
		}
		if serviceAccountMatches(sts.Spec.Template.Spec.ServiceAccountName, n) {
			refs = append(refs, Ref{
				GVR: s.GVR(),
				FQN: client.FQN(sts.Namespace, sts.Name),
			})
		}
	}

	return refs, nil
}

// Scan scans for cluster resource refs.
func (s *StatefulSet) Scan(ctx context.Context, gvr client.GVR, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := s.getFactory().List(s.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var sts appsv1.StatefulSet
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sts)
		if err != nil {
			return nil, errors.New("expecting StatefulSet resource")
		}
		switch gvr {
		case CmGVR:
			if !hasConfigMap(&sts.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: s.GVR(),
				FQN: client.FQN(sts.Namespace, sts.Name),
			})
		case SecGVR:
			found, err := hasSecret(s.Factory, &sts.Spec.Template.Spec, sts.Namespace, n, wait)
			if err != nil {
				log.Warn().Err(err).Msgf("locate secret %q", fqn)
				continue
			}
			if !found {
				continue
			}
			refs = append(refs, Ref{
				GVR: s.GVR(),
				FQN: client.FQN(sts.Namespace, sts.Name),
			})
		case PvcGVR:
			for _, v := range sts.Spec.VolumeClaimTemplates {
				if !strings.HasPrefix(n, v.Name+"-"+sts.Name) {
					continue
				}
				refs = append(refs, Ref{
					GVR: s.GVR(),
					FQN: client.FQN(sts.Namespace, sts.Name),
				})
			}
			if !hasPVC(&sts.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: s.GVR(),
				FQN: client.FQN(sts.Namespace, sts.Name),
			})
		case PcGVR:
			if !hasPC(&sts.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: s.GVR(),
				FQN: client.FQN(sts.Namespace, sts.Name),
			})

		}
	}

	return refs, nil
}

// GetPodSpec returns a pod spec given a resource.
func (s *StatefulSet) GetPodSpec(path string) (*v1.PodSpec, error) {
	sts, err := s.getStatefulSet(path)
	if err != nil {
		return nil, err
	}
	podSpec := sts.Spec.Template.Spec
	return &podSpec, nil
}

// SetImages sets container images.
func (s *StatefulSet) SetImages(ctx context.Context, path string, imageSpecs ImageSpecs) error {
	ns, n := client.Namespaced(path)
	auth, err := s.Client().CanI(ns, "apps/v1/statefulset", n, client.PatchAccess)
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to patch a statefulset")
	}
	jsonPatch, err := GetTemplateJsonPatch(imageSpecs)
	if err != nil {
		return err
	}
	dial, err := s.Client().Dial()
	if err != nil {
		return err
	}
	_, err = dial.AppsV1().StatefulSets(ns).Patch(
		ctx,
		n,
		types.StrategicMergePatchType,
		jsonPatch,
		metav1.PatchOptions{},
	)
	return err
}

func podsFromSelector(f Factory, ns string, sel map[string]string) ([]*v1.Pod, error) {
	oo, err := f.List("v1/pods", ns, true, labels.Set(sel).AsSelector())
	if err != nil {
		return nil, err
	}

	if len(oo) == 0 {
		return nil, fmt.Errorf("no matching pods for %v", sel)
	}

	pp := make([]*v1.Pod, 0, len(oo))
	for _, o := range oo {
		pod := new(v1.Pod)
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, pod)
		if err != nil {
			return nil, err
		}
		pp = append(pp, pod)
	}

	return pp, nil
}
