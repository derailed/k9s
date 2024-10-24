// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tview"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// DaemonSet renders a K8s DaemonSet to screen.
type DaemonSet struct {
	Base
}

// Header returns a header row.
func (DaemonSet) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "VS", VS: true},
		model1.HeaderColumn{Name: "DESIRED", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "CURRENT", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "READY", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "UP-TO-DATE", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "AVAILABLE", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "LABELS", Wide: true},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (d DaemonSet) Render(o interface{}, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected DaemonSet, but got %T", o)
	}
	var ds appsv1.DaemonSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ds)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(ds.ObjectMeta)
	r.Fields = model1.Fields{
		ds.Namespace,
		ds.Name,
		computeVulScore(ds.ObjectMeta, &ds.Spec.Template.Spec),
		strconv.Itoa(int(ds.Status.DesiredNumberScheduled)),
		strconv.Itoa(int(ds.Status.CurrentNumberScheduled)),
		strconv.Itoa(int(ds.Status.NumberReady)),
		strconv.Itoa(int(ds.Status.UpdatedNumberScheduled)),
		strconv.Itoa(int(ds.Status.NumberAvailable)),
		mapToStr(ds.Labels),
		AsStatus(d.diagnose(ds.Status.DesiredNumberScheduled, ds.Status.NumberReady)),
		ToAge(ds.GetCreationTimestamp()),
	}

	return nil
}

// Happy returns true if resource is happy, false otherwise.
func (DaemonSet) diagnose(d, r int32) error {
	if d != r {
		return fmt.Errorf("desiring %d replicas but %d ready", d, r)
	}
	return nil
}
