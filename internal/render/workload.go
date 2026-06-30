// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var defaultWKHeader = model1.Header{
	model1.HeaderColumn{Name: "KIND"},
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "STATUS"},
	model1.HeaderColumn{Name: "READY"},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// Workload renders a workload to screen.
type Workload struct {
	Base
}

// wkKindColors tints workload rows by KIND so different resource types are
// visually grouped: rows sharing a KIND share a color. Red is intentionally
// omitted to avoid clashing with the DEGRADED warning color.
var wkKindColors = []tcell.Color{
	tcell.ColorTeal,
	tcell.ColorNavy,
	tcell.ColorOlive,
	tcell.ColorPurple,
	tcell.ColorGreen,
	tcell.ColorBlue,
	tcell.ColorFuchsia,
	tcell.ColorAqua,
}

// ColorerFunc colors a resource row. DEGRADED status takes precedence;
// otherwise each row is tinted by its KIND to visually separate the different
// resource types aggregated in the workload view.
func (Workload) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		if idx, ok := h.IndexOf("STATUS", true); ok {
			if strings.TrimSpace(re.Row.Fields[idx]) == "DEGRADED" {
				return model1.PendingColor
			}
		}
		if idx, ok := h.IndexOf("KIND", true); ok {
			hsh := fnv.New32a()
			_, _ = hsh.Write([]byte(re.Row.Fields[idx]))
			return wkKindColors[hsh.Sum32()%uint32(len(wkKindColors))]
		}

		return model1.DefaultColorer(ns, h, re)
	}
}

// Header returns a header rbw.
func (Workload) Header(string) model1.Header {
	return defaultWKHeader
}

// Render renders a K8s resource to screen.
func (Workload) Render(o any, _ string, r *model1.Row) error {
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

	return nil
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
