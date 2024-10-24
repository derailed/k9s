// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Workload renders a workload to screen.
type Workload struct {
	Base
}

// ColorerFunc colors a resource row.
func (n Workload) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)

		idx, ok := h.IndexOf("STATUS", true)
		if !ok {
			return c
		}
		status := strings.TrimSpace(re.Row.Fields[idx])
		if status == "DEGRADED" {
			c = model1.PendingColor
		}

		return c
	}
}

// Header returns a header rbw.
func (Workload) Header(string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "KIND"},
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "STATUS"},
		model1.HeaderColumn{Name: "READY"},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (n Workload) Render(o interface{}, _ string, r *model1.Row) error {
	res, ok := o.(*WorkloadRes)
	if !ok {
		return fmt.Errorf("expected allRes but got %T", o)
	}

	r.ID = fmt.Sprintf("%s|%s|%s", res.Row.Cells[0].(string), res.Row.Cells[1].(string), res.Row.Cells[2].(string))
	r.Fields = model1.Fields{
		res.Row.Cells[0].(string),
		res.Row.Cells[1].(string),
		res.Row.Cells[2].(string),
		res.Row.Cells[3].(string),
		res.Row.Cells[4].(string),
		res.Row.Cells[5].(string),
		ToAge(res.Row.Cells[6].(metav1.Time)),
	}

	return nil
}

type WorkloadRes struct {
	Row metav1.TableRow
}

// GetObjectKind returns a schema object.
func (a *WorkloadRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (a *WorkloadRes) DeepCopyObject() runtime.Object {
	return a
}
