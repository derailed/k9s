// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
)

var (
	_ Scalable       = (*Scaler)(nil)
	_ ReplicasGetter = (*Scaler)(nil)
)

// Scaler represents a generic resource with scaling.
type Scaler struct {
	Generic
}

// Replicas returns the number of replicas for the resource located at the given path.
func (s *Scaler) Replicas(ctx context.Context, path string) (int32, error) {
	scaleClient, err := s.scaleClient()
	if err != nil {
		return 0, err
	}

	ns, name := client.Namespaced(path)
	currScale, err := scaleClient.Scales(ns).Get(ctx, *s.gvr.GR(), name, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	return currScale.Spec.Replicas, nil
}

// Scale modifies the number of replicas for a given resource specified by the path.
func (s *Scaler) Scale(ctx context.Context, path string, replicas int32) error {
	ns, name := client.Namespaced(path)

	scaleClient, err := s.scaleClient()
	if err != nil {
		return err
	}

	scales := scaleClient.Scales(ns)
	patch := []byte(fmt.Sprintf(`{"spec":{"replicas":%d}}`, replicas))
	updatedScale, err := scales.Patch(ctx, s.gvr.GVR(), name, types.MergePatchType, patch, metav1.PatchOptions{})
	if kerrors.IsForbidden(err) {
		// Roles granting get+update but not patch on the scale subresource keep working.
		updatedScale, err = s.getAndUpdate(ctx, scales, name, replicas)
	}
	if err != nil {
		return err
	}

	slog.Debug("Scaled resource",
		slogs.FQN, path,
		slogs.Replicas, updatedScale.Spec.Replicas,
	)
	return nil
}

func (s *Scaler) getAndUpdate(ctx context.Context, scales scale.ScaleInterface, name string, replicas int32) (*autoscalingv1.Scale, error) {
	currentScale, err := scales.Get(ctx, *s.gvr.GR(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	currentScale.Spec.Replicas = replicas

	return scales.Update(ctx, *s.gvr.GR(), currentScale, metav1.UpdateOptions{})
}

func (s *Scaler) scaleClient() (scale.ScalesGetter, error) {
	cfg, err := s.Client().RestConfig()
	if err != nil {
		return nil, err
	}

	discoveryClient, err := s.Client().CachedDiscovery()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	scaleKindResolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)

	return scale.NewForConfig(cfg, mapper, dynamic.LegacyAPIPathResolverFunc, scaleKindResolver)
}
