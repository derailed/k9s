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

var defaultSTSHeader = model1.Header{
	model1.HeaderColumn{Name: colNamespace},
	model1.HeaderColumn{Name: colName},
	model1.HeaderColumn{Name: colVS, Attrs: model1.Attrs{VS: true}},
	model1.HeaderColumn{Name: colReady},
	model1.HeaderColumn{Name: colSelector, Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "SERVICE"},
	model1.HeaderColumn{Name: colContainers, Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: colImages, Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: colLabels, Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: colValid, Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: colAge, Attrs: model1.Attrs{Time: true}},
}

// StatefulSet renders a K8s StatefulSet to screen.
type StatefulSet struct {
	Base
}

// Header returns a header row.
func (s StatefulSet) Header(_ string) model1.Header {
	return s.doHeader(defaultSTSHeader)
}

// Render renders a K8s resource to screen.
func (s StatefulSet) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := s.defaultRow(raw, row); err != nil {
		return err
	}
	if s.specs.isEmpty() {
		return nil
	}
	cols, err := s.specs.realize(raw, defaultSTSHeader, row)
	cols.hydrateRow(row)

	return err
}

func (s StatefulSet) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	var sts appsv1.StatefulSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sts)
	if err != nil {
		return err
	}

	var desired int32
	if sts.Spec.Replicas != nil {
		desired = *sts.Spec.Replicas
	}
	r.ID = client.MetaFQN(&sts.ObjectMeta)
	r.Fields = model1.Fields{
		sts.Namespace,
		sts.Name,
		computeVulScore(sts.Namespace, sts.Labels, &sts.Spec.Template.Spec),
		strconv.Itoa(int(sts.Status.ReadyReplicas)) + "/" + strconv.Itoa(int(desired)),
		asSelector(sts.Spec.Selector),
		na(sts.Spec.ServiceName),
		podContainerNames(&sts.Spec.Template.Spec, true),
		podImageNames(&sts.Spec.Template.Spec, true),
		mapToStr(sts.Labels),
		AsStatus(s.diagnose(desired, sts.Status.Replicas, sts.Status.ReadyReplicas)),
		ToAge(sts.GetCreationTimestamp()),
	}

	return nil
}

func (StatefulSet) diagnose(d, c, r int32) error {
	if c != r {
		return fmt.Errorf("desired %d replicas got %d available", c, r)
	}
	if d != r {
		return fmt.Errorf("want %d replicas got %d available", d, r)
	}

	return nil
}
