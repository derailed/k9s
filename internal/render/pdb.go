// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tview"
	v1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PodDisruptionBudget renders a K8s PodDisruptionBudget to screen.
type PodDisruptionBudget struct {
	Base
}

// Header returns a header row.
func (PodDisruptionBudget) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "MIN AVAILABLE", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "MAX UNAVAILABLE", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "ALLOWED DISRUPTIONS", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "CURRENT", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "DESIRED", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "EXPECTED", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "LABELS", Wide: true},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (p PodDisruptionBudget) Render(o interface{}, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected PodDisruptionBudget, but got %T", o)
	}
	var pdb v1.PodDisruptionBudget
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &pdb)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(pdb.ObjectMeta)
	r.Fields = model1.Fields{
		pdb.Namespace,
		pdb.Name,
		numbToStr(pdb.Spec.MinAvailable),
		numbToStr(pdb.Spec.MaxUnavailable),
		strconv.Itoa(int(pdb.Status.DisruptionsAllowed)),
		strconv.Itoa(int(pdb.Status.CurrentHealthy)),
		strconv.Itoa(int(pdb.Status.DesiredHealthy)),
		strconv.Itoa(int(pdb.Status.ExpectedPods)),
		mapToStr(pdb.Labels),
		AsStatus(p.diagnose(pdb.Spec.MinAvailable, pdb.Status.CurrentHealthy)),
		ToAge(pdb.GetCreationTimestamp()),
	}

	return nil
}

func (PodDisruptionBudget) diagnose(min *intstr.IntOrString, healthy int32) error {
	if min == nil {
		return nil
	}
	if min.IntVal > healthy {
		return fmt.Errorf("expected %d but got %d", min.IntVal, healthy)
	}
	return nil
}

// Helpers...

func numbToStr(n *intstr.IntOrString) string {
	if n == nil {
		return NAValue
	}
	if n.Type == intstr.Int {
		return strconv.Itoa(int(n.IntVal))
	}
	return n.StrVal
}
