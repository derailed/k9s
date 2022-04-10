package render

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
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
	return getColumns()
}

func getColumns() []HeaderColumn {
	columns := []HeaderColumn{
		{Name: "RESOURCE"},
		{Name: "TOTAL"},
		{Name: "MODIFIED", Hide: true},
		{Name: "ADDED", Hide: true},
		{Name: "PENDING", Hide: true},
		{Name: "ERROR", Hide: true},
		{Name: "STD", Hide: true},
		{Name: "HIGHLIGHT", Hide: true},
		{Name: "KILL", Hide: true},
		{Name: "COMPLETED", Hide: true},
	}

	columnsMap := make(map[string]HeaderColumn)
	dashboardsConfig, err := config.LoadDashboard()
	if err == nil {
		for _, dashboardConfig := range dashboardsConfig.GVRs {
			if dashboardConfig.Active {
				for column := range dashboardConfig.Columns {
					columnsMap[column] = HeaderColumn{Name: column, Wide: true}
				}
			}
		}
	}

	customColumns := make([]HeaderColumn, 0)
	for _, column := range columnsMap {
		customColumns = append(customColumns, column)
	}
	sort.Slice(customColumns, func(i, j int) bool {
		return customColumns[i].Name < customColumns[j].Name
	})

	for _, column := range customColumns {
		columns = append(columns, column)
	}

	return columns
}

type DashboardRes struct {
	GVR    client.GVR
	Counts map[string]int
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
	r.Fields = append(r.Fields, dash.GVR.String())

	for _, column := range getColumns()[1:] {
		r.Fields = append(r.Fields, strconv.Itoa(dash.Counts[column.Name]))
	}

	return nil
}

func (Dashboard) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		modified, _ := strconv.Atoi(getField("MODIFIED", h, re))
		added, _ := strconv.Atoi(getField("ADDED", h, re))
		pending, _ := strconv.Atoi(getField("PENDING", h, re))
		errors, _ := strconv.Atoi(getField("ERROR", h, re))
		std, _ := strconv.Atoi(getField("STD", h, re))
		highlight, _ := strconv.Atoi(getField("HIGHLIGHT", h, re))
		kill, _ := strconv.Atoi(getField("KILL", h, re))
		completed, _ := strconv.Atoi(getField("COMPLETED", h, re))

		switch {
		case errors > 0:
			return ErrColor
		case highlight > 0:
			return HighlightColor
		case kill > 0:
			return KillColor
		case pending > 0:
			return PendingColor
		case modified > 0:
			return ModColor
		case added > 0:
			return AddColor
		case completed > 0:
			return CompletedColor
		case std > 0:
			return StdColor
		default:
			return StdColor
		}
	}
}

func getField(name string, h Header, re RowEvent) string {
	idx := h.IndexOf(name, true)
	if idx == -1 {
		return ""
	}
	return re.Row.Fields[idx]
}
