// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

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
func (n Workload) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		c := DefaultColorer(ns, h, re)

		statusCol := h.IndexOf("STATUS", true)
		if statusCol == -1 {
			return c
		}
		status := strings.TrimSpace(re.Row.Fields[statusCol])
		if status == "DEGRADED" {
			c = PendingColor
		}

		return c
	}
}

// Header returns a header rbw.
func (Workload) Header(string) Header {
	return Header{
		HeaderColumn{Name: "KIND"},
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "STATUS"},
		HeaderColumn{Name: "READY"},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (n Workload) Render(o interface{}, _ string, r *Row) error {
	res, ok := o.(*WorkloadRes)
	if !ok {
		return fmt.Errorf("expected allRes but got %T", o)
	}

	r.ID = fmt.Sprintf("%s|%s|%s", res.Row.Cells[0].(string), res.Row.Cells[1].(string), res.Row.Cells[2].(string))
	r.Fields = Fields{
		res.Row.Cells[0].(string),
		res.Row.Cells[1].(string),
		res.Row.Cells[2].(string),
		res.Row.Cells[3].(string),
		res.Row.Cells[4].(string),
		ToAge(res.Row.Cells[5].(metav1.Time)),
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
