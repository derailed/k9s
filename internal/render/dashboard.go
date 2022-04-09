package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Dashboard renders a dashboard to screen.
type Dashboard struct {
	Base
}

// Header returns a header row.
func (Dashboard) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "RESOURCE"},
		HeaderColumn{Name: "TOTAL"},
		HeaderColumn{Name: "OK"},
		HeaderColumn{Name: "ERRORS"},
	}
}

type DashboardRes struct {
	GVR    client.GVR
	Total  int
	OK     int
	Errors int
}

// GetObjectKind returns a schema object.
func (DashboardRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (a DashboardRes) DeepCopyObject() runtime.Object {
	return a
}

// Render renders a K8s resource to screen.
func (Dashboard) Render(o interface{}, ns string, r *Row) error {
	dash, ok := o.(DashboardRes)
	if !ok {
		return fmt.Errorf("expected DashboardRes, but got %T", o)
	}

	r.ID = dash.GVR.String()
	r.Fields = append(r.Fields,
		dash.GVR.String(),
		strconv.Itoa(dash.Total),
		strconv.Itoa(dash.OK),
		strconv.Itoa(dash.Errors),
	)
	return nil
}

func (Dashboard) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		total := getField("TOTAL", h, re)
		ok := getField("OK", h, re)
		// errors := getField("ERRORS", h, re)

		if total == ok {
			return ModColor
		}

		return ErrColor
	}
}

func getField(name string, h Header, re RowEvent) string {
	idx := h.IndexOf(name, true)
	if idx == -1 {
		return ""
	}
	return re.Row.Fields[idx]
}
