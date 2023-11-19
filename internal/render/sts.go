// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// StatefulSet renders a K8s StatefulSet to screen.
type StatefulSet struct {
	Base
}

// Header returns a header row.
func (StatefulSet) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "READY"},
		HeaderColumn{Name: "SELECTOR", Wide: true},
		HeaderColumn{Name: "SERVICE"},
		HeaderColumn{Name: "CONTAINERS", Wide: true},
		HeaderColumn{Name: "IMAGES", Wide: true},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (s StatefulSet) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected StatefulSet, but got %T", o)
	}
	var sts appsv1.StatefulSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sts)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(sts.ObjectMeta)
	r.Fields = Fields{
		sts.Namespace,
		sts.Name,
		strconv.Itoa(int(sts.Status.ReadyReplicas)) + "/" + strconv.Itoa(int(sts.Status.Replicas)),
		asSelector(sts.Spec.Selector),
		na(sts.Spec.ServiceName),
		podContainerNames(sts.Spec.Template.Spec, true),
		podImageNames(sts.Spec.Template.Spec, true),
		mapToStr(sts.Labels),
		asStatus(s.diagnose(sts.Status.Replicas, sts.Status.ReadyReplicas)),
		toAge(sts.GetCreationTimestamp()),
	}

	return nil
}

func (StatefulSet) diagnose(d, r int32) error {
	if d != r {
		return fmt.Errorf("desiring %d replicas got %d available", d, r)
	}
	return nil
}
