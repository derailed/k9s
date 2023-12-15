// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
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
func (d Deployment) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		c := DefaultColorer(ns, h, re)
		if !Happy(ns, h, re.Row) {
			return ErrColor
		}
		rdCol := h.IndexOf("READY", true)
		if rdCol == -1 {
			return c
		}
		ready := strings.TrimSpace(re.Row.Fields[rdCol])
		tt := strings.Split(ready, "/")
		if len(tt) == 2 && tt[1] == "0" {
			return PendingColor
		}

		return c
	}
}

// Header returns a header row.
func (Deployment) Header(ns string) Header {
	h := Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "VS", VS: true},
		HeaderColumn{Name: "READY", Align: tview.AlignRight},
		HeaderColumn{Name: "UP-TO-DATE", Align: tview.AlignRight},
		HeaderColumn{Name: "AVAILABLE", Align: tview.AlignRight},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}

	return h
}

// Render renders a K8s resource to screen.
func (d Deployment) Render(o interface{}, ns string, r *Row) error {
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
	r.Fields = Fields{
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
