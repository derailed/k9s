package dao

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*Service)(nil)
	_ Loggable = (*Service)(nil)
)

// Service represents a k8s service.
type Service struct {
	Resource
}

// TailLogs tail logs for all pods represented by this Service.
func (s *Service) TailLogs(ctx context.Context, c chan<- []byte, opts LogOptions) error {
	o, err := s.Factory.Get(s.gvr.String(), opts.Path, true, labels.Everything())
	if err != nil {
		return err
	}
	var svc v1.Service
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &svc)
	if err != nil {
		return errors.New("expecting Service resource")
	}

	if svc.Spec.Selector == nil || len(svc.Spec.Selector) == 0 {
		return fmt.Errorf("no valid selector found on Service %s", opts.Path)
	}

	return podLogs(ctx, c, svc.Spec.Selector, opts)
}
