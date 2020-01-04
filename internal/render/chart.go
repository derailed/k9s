package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Chart renders a helm chart to screen.
type Chart struct{}

// ColorerFunc colors a resource row.
func (Chart) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Chart) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "REVISION"},
		Header{Name: "STATUS"},
		Header{Name: "CHART"},
		Header{Name: "APP VERSION"},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a chart to screen.
func (c Chart) Render(o interface{}, ns string, r *Row) error {
	h, ok := o.(ChartRes)
	if !ok {
		return fmt.Errorf("expected ChartRes, but got %T", o)
	}

	r.ID = FQN(h.Release.Namespace, h.Release.Name)
	r.Fields = make(Fields, 0, len(c.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, h.Release.Namespace)
	}
	r.Fields = append(r.Fields,
		h.Release.Name,
		strconv.Itoa(h.Release.Version),
		h.Release.Info.Status.String(),
		h.Release.Chart.Metadata.Name+"-"+h.Release.Chart.Metadata.Version,
		h.Release.Chart.Metadata.AppVersion,
		toAge(metav1.Time{Time: h.Release.Info.LastDeployed.Time}),
	)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// ChartRes represents an alias resource.
type ChartRes struct {
	Release *release.Release
}

// GetObjectKind returns a schema object.
func (ChartRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (h ChartRes) DeepCopyObject() runtime.Object {
	return h
}
