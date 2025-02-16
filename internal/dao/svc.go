// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor   = (*Service)(nil)
	_ Loggable   = (*Service)(nil)
	_ Controller = (*Service)(nil)
)

// Service represents a k8s service.
type Service struct {
	Resource
}

// TailLogs tail logs for all pods represented by this Service.
func (s *Service) TailLogs(ctx context.Context, opts *LogOptions) ([]LogChan, error) {
	svc, err := s.GetInstance(opts.Path)
	if err != nil {
		return nil, err
	}
	if len(svc.Spec.Selector) == 0 {
		return nil, fmt.Errorf("no valid selector found on Service %s", opts.Path)
	}

	return podLogs(ctx, svc.Spec.Selector, opts)
}

// Pod returns a pod victim by name.
func (s *Service) Pod(fqn string) (string, error) {
	svc, err := s.GetInstance(fqn)
	if err != nil {
		return "", err
	}

	return podFromSelector(s.Factory, svc.Namespace, svc.Spec.Selector)
}

// GetInstance returns a service instance.
func (s *Service) GetInstance(fqn string) (*v1.Service, error) {
	o, err := s.getFactory().Get(s.gvrStr(), fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var svc v1.Service
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &svc)
	if err != nil {
		return nil, errors.New("expecting Service resource")
	}

	return &svc, nil
}

// ----------------------------------------------------------------------------
// Helpers...

func podFromSelector(f Factory, ns string, sel map[string]string) (string, error) {
	oo, err := f.List("v1/pods", ns, true, labels.Set(sel).AsSelector())
	if err != nil {
		return "", err
	}

	if len(oo) == 0 {
		return "", fmt.Errorf("no matching pods for %v", sel)
	}

	var pod v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(oo[0].(*unstructured.Unstructured).Object, &pod)
	if err != nil {
		return "", err
	}

	return client.FQN(pod.Namespace, pod.Name), nil
}
