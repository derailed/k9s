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
	"k8s.io/apimachinery/pkg/util/sets"
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
	ns, n := client.Namespaced(opts.Path)
	pods, err := podsForService(s.Factory, ns, n)
	if err != nil {
		return nil, err
	}
	if pods.Len() == 0 {
		return nil, fmt.Errorf("no pods backing Service %s", opts.Path)
	}

	return tailPodsLogs(ctx, s.Factory, sets.List(pods), opts)
}

// Pod returns a pod victim by name.
func (s *Service) Pod(fqn string) (string, error) {
	ns, n := client.Namespaced(fqn)
	pods, err := podsForService(s.Factory, ns, n)
	if err != nil {
		return "", err
	}
	if pods.Len() == 0 {
		return "", fmt.Errorf("no matching pods for Service %s", fqn)
	}

	return sets.List(pods)[0], nil
}

// GetInstance returns a service instance.
func (s *Service) GetInstance(fqn string) (*v1.Service, error) {
	o, err := s.getFactory().Get(s.gvr, fqn, true, labels.Everything())
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
