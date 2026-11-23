// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var defaultWKHeader = model1.Header{
	model1.HeaderColumn{Name: colKind},
	model1.HeaderColumn{Name: colNamespace},
	model1.HeaderColumn{Name: colName},
	model1.HeaderColumn{Name: colStatus},
	model1.HeaderColumn{Name: colReady},
	model1.HeaderColumn{Name: colValid, Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: colAge, Attrs: model1.Attrs{Time: true}},
}

// Workload renders a workload to screen.
type Workload struct {
	Base
}

// ColorerFunc colors a resource row.
func (Workload) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)

		switch workloadRowKind(h, re) {
		case client.DpGVR.String():
			if isScaledToZero(h, re) {
				return model1.PendingColor
			}
			return c
		case client.StsGVR.String():
			return c
		}

		idx, ok := h.IndexOf(colStatus, true)
		if !ok || idx >= len(re.Row.Fields) {
			return c
		}
		status := strings.TrimSpace(re.Row.Fields[idx])
		if status == "DEGRADED" {
			c = model1.PendingColor
		}

		return c
	}
}

func workloadRowKind(h model1.Header, re *model1.RowEvent) string {
	if gvr, _, ok := strings.Cut(re.Row.ID, "|"); ok && gvr != "" {
		return gvr
	}
	if idx, ok := h.IndexOf(colKind, true); ok && idx < len(re.Row.Fields) {
		return strings.TrimSpace(re.Row.Fields[idx])
	}

	return ""
}

func isScaledToZero(h model1.Header, re *model1.RowEvent) bool {
	idx, ok := h.IndexOf(colReady, true)
	if !ok || idx >= len(re.Row.Fields) {
		return false
	}
	ready := strings.TrimSpace(re.Row.Fields[idx])
	tt := strings.Split(ready, "/")

	return len(tt) == 2 && tt[1] == "0"
}

// Header returns a header row.
func (w Workload) Header(string) model1.Header {
	return w.doHeader(defaultWKHeader)
}

// Render renders a K8s resource to screen.
func (w Workload) Render(o any, _ string, r *model1.Row) error {
	res, ok := o.(*WorkloadRes)
	if !ok {
		return fmt.Errorf("expected WorkloadRes but got %T", o)
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
	stashAge(r, res.Row.Cells[6].(metav1.Time))
	if w.specs.isEmpty() {
		return nil
	}

	obj := res.Row.Object.Object
	if obj != nil {
		obj = obj.DeepCopyObject()
	}
	cols, err := w.specs.realize(obj, defaultWKHeader, r)
	cols.hydrateRow(r)

	return err
}

type WorkloadRes struct {
	Row metav1.TableRow
}

// GetObjectKind returns a schema object.
func (*WorkloadRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (a *WorkloadRes) DeepCopyObject() runtime.Object {
	return a
}
