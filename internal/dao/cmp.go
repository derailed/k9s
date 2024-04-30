// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"errors"

	appsvalpha1 "github.com/apecloud/kubeblocks/apis/apps/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*Cluster)(nil)
	_ Nuker    = (*Cluster)(nil)
)

// Cluster represents a K8s sts.
type Cluster struct {
	Resource
}

// GetInstance returns a statefulset instance.
func (*Cluster) GetInstance(f Factory, fqn string) (*appsvalpha1.Cluster, error) {
	o, err := f.Get("apps.kubeblocks.io/v1alpha1/clusters", fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var cmp appsvalpha1.Cluster
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cmp)
	if err != nil {
		return nil, errors.New("expecting Statefulset resource")
	}

	return &cmp, nil
}

func (s *Cluster) getIinstance(fqn string) (*appsvalpha1.Cluster, error) {
	o, err := s.getFactory().Get(s.gvrStr(), fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var sts appsvalpha1.Cluster
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sts)
	if err != nil {
		return nil, errors.New("expecting Service resource")
	}

	return &sts, nil
}
