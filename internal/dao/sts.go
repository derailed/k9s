// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

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
func (s *StatefulSet) ListImages(_ context.Context, fqn string) ([]string, error) {
	sts, err := s.GetInstance(s.Factory, fqn)
	if err != nil {
		return nil, err
	}

	return render.ExtractImages(&sts.Spec.Template.Spec), nil
}

// Scale a StatefulSet.
func (s *StatefulSet) Scale(ctx context.Context, path string, replicas int32) error {
	return scaleRes(ctx, s.getFactory(), client.StsGVR, path, replicas)
}

// Restart a StatefulSet rollout.
func (s *StatefulSet) Restart(ctx context.Context, path string) error {
	return restartRes[*appsv1.StatefulSet](ctx, s.getFactory(), client.StsGVR, path)
}

// GetInstance returns a statefulset instance.
func (*StatefulSet) GetInstance(f Factory, fqn string) (*appsv1.StatefulSet, error) {
	o, err := f.Get(client.StsGVR, fqn, true, labels.Everything())
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
	o, err := s.getFactory().Get(s.gvr, fqn, true, labels.Everything())
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
func (s *StatefulSet) ScanSA(_ context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := s.getFactory().List(s.gvr, ns, wait, labels.Everything())
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
func (s *StatefulSet) Scan(_ context.Context, gvr *client.GVR, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := s.getFactory().List(s.gvr, ns, wait, labels.Everything())
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
		case client.CmGVR:
			if !hasConfigMap(&sts.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: s.GVR(),
				FQN: client.FQN(sts.Namespace, sts.Name),
			})
		case client.SecGVR:
			found, err := hasSecret(s.Factory, &sts.Spec.Template.Spec, sts.Namespace, n, wait)
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
				GVR: s.GVR(),
				FQN: client.FQN(sts.Namespace, sts.Name),
			})
		case client.PvcGVR:
			for i := range sts.Spec.VolumeClaimTemplates {
				if !strings.HasPrefix(n, sts.Spec.VolumeClaimTemplates[i].Name+"-"+sts.Name) {
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
		case client.PcGVR:
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
	auth, err := s.Client().CanI(ns, client.StsGVR, n, client.PatchAccess)
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
