// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Deployment renders a K8s Deployment to screen.
type Deployment struct {
	Base
}

// ColorerFunc colors a resource row.
func (d Deployment) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)

		idx, ok := h.IndexOf("READY", true)
		if !ok {
			return c
		}
		ready := strings.TrimSpace(re.Row.Fields[idx])
		tt := strings.Split(ready, "/")
		if len(tt) == 2 && tt[1] == "0" {
			return model1.PendingColor
		}

		return c
	}
}

// Header returns a header row.
func (Deployment) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "VS", VS: true},
		model1.HeaderColumn{Name: "READY", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "UP-TO-DATE", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "AVAILABLE", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "LABELS", Wide: true},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (d Deployment) Render(o interface{}, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Deployment, but got %T", o)
	}

	var dp appsv1.Deployment
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &dp)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(dp.ObjectMeta)
	r.Fields = model1.Fields{
		dp.Namespace,
		dp.Name,
		computeVulScore(dp.ObjectMeta, &dp.Spec.Template.Spec),
		strconv.Itoa(int(dp.Status.AvailableReplicas)) + "/" + strconv.Itoa(int(dp.Status.Replicas)),
		strconv.Itoa(int(dp.Status.UpdatedReplicas)),
		strconv.Itoa(int(dp.Status.AvailableReplicas)),
		mapToStr(dp.Labels),
		AsStatus(d.diagnose(dp.Status.Replicas, dp.Status.AvailableReplicas)),
		ToAge(dp.GetCreationTimestamp()),
	}

	return nil
}

func (Deployment) diagnose(desired, avail int32) error {
	if desired != avail {
		return fmt.Errorf("desiring %d replicas got %d available", desired, avail)
	}

	return nil
}
