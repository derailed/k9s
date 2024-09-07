// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// StatefulSet renders a K8s StatefulSet to screen.
type StatefulSet struct {
	Base
}

// Header returns a header row.
func (StatefulSet) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "VS", VS: true},
		model1.HeaderColumn{Name: "READY"},
		model1.HeaderColumn{Name: "SELECTOR", Wide: true},
		model1.HeaderColumn{Name: "SERVICE"},
		model1.HeaderColumn{Name: "CONTAINERS", Wide: true},
		model1.HeaderColumn{Name: "IMAGES", Wide: true},
		model1.HeaderColumn{Name: "LABELS", Wide: true},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (s StatefulSet) Render(o interface{}, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected StatefulSet, but got %T", o)
	}
	var sts appsv1.StatefulSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sts)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(sts.ObjectMeta)
	r.Fields = model1.Fields{
		sts.Namespace,
		sts.Name,
		computeVulScore(sts.ObjectMeta, &sts.Spec.Template.Spec),
		strconv.Itoa(int(sts.Status.ReadyReplicas)) + "/" + strconv.Itoa(int(sts.Status.Replicas)),
		asSelector(sts.Spec.Selector),
		na(sts.Spec.ServiceName),
		podContainerNames(sts.Spec.Template.Spec, true),
		podImageNames(sts.Spec.Template.Spec, true),
		mapToStr(sts.Labels),
		AsStatus(s.diagnose(sts.Spec.Replicas, sts.Status.Replicas, sts.Status.ReadyReplicas)),
		ToAge(sts.GetCreationTimestamp()),
	}

	return nil
}

func (StatefulSet) diagnose(w *int32, d, r int32) error {
	if d != r {
		return fmt.Errorf("desired %d replicas got %d available", d, r)
	}
	if w != nil && *w != r {
		return fmt.Errorf("want %d replicas got %d available", *w, r)
	}

	return nil
}
