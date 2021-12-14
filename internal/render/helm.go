package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell/v2"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Helm renders a helm chart to screen.
type Helm struct{}

// IsGeneric identifies a generic handler.
func (Helm) IsGeneric() bool {
	return false
}

// ColorerFunc colors a resource row.
func (Helm) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		if !Happy(ns, h, re.Row) {
			return ErrColor
		}

		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Helm) Header(_ string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "REVISION"},
		HeaderColumn{Name: "STATUS"},
		HeaderColumn{Name: "CHART"},
		HeaderColumn{Name: "APP VERSION"},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true, Decorator: AgeDecorator},
	}
}

// Render renders a chart to screen.
func (c Helm) Render(o interface{}, ns string, r *Row) error {
	h, ok := o.(HelmRes)
	if !ok {
		return fmt.Errorf("expected HelmRes, but got %T", o)
	}

	r.ID = client.FQN(h.Release.Namespace, h.Release.Name)
	r.Fields = Fields{
		h.Release.Namespace,
		h.Release.Name,
		strconv.Itoa(h.Release.Version),
		h.Release.Info.Status.String(),
		h.Release.Chart.Metadata.Name + "-" + h.Release.Chart.Metadata.Version,
		h.Release.Chart.Metadata.AppVersion,
		asStatus(c.diagnose(h.Release.Info.Status.String())),
		toAge(metav1.Time{Time: h.Release.Info.LastDeployed.Time}),
	}

	return nil
}

func (c Helm) diagnose(s string) error {
	if s != "deployed" {
		return fmt.Errorf("chart is in an invalid state")
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// HelmRes represents an helm chart resource.
type HelmRes struct {
	Release *release.Release
}

// GetObjectKind returns a schema object.
func (HelmRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (h HelmRes) DeepCopyObject() runtime.Object {
	return h
}
