// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package helm

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Chart renders a helm chart to screen.
type Chart struct{}

// IsGeneric identifies a generic handler.
func (Chart) IsGeneric() bool {
	return false
}

// ColorerFunc colors a resource row.
func (Chart) ColorerFunc() render.ColorerFunc {
	return render.DefaultColorer
}

// Header returns a header row.
func (Chart) Header(_ string) render.Header {
	return render.Header{
		render.HeaderColumn{Name: "NAMESPACE"},
		render.HeaderColumn{Name: "NAME"},
		render.HeaderColumn{Name: "REVISION"},
		render.HeaderColumn{Name: "STATUS"},
		render.HeaderColumn{Name: "CHART"},
		render.HeaderColumn{Name: "APP VERSION"},
		render.HeaderColumn{Name: "VALID", Wide: true},
		render.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a chart to screen.
func (c Chart) Render(o interface{}, ns string, r *render.Row) error {
	h, ok := o.(ReleaseRes)
	if !ok {
		return fmt.Errorf("expected ReleaseRes, but got %T", o)
	}

	r.ID = client.FQN(h.Release.Namespace, h.Release.Name)
	r.Fields = render.Fields{
		h.Release.Namespace,
		h.Release.Name,
		strconv.Itoa(h.Release.Version),
		h.Release.Info.Status.String(),
		h.Release.Chart.Metadata.Name + "-" + h.Release.Chart.Metadata.Version,
		h.Release.Chart.Metadata.AppVersion,
		render.AsStatus(c.diagnose(h.Release.Info.Status.String())),
		render.ToAge(metav1.Time{Time: h.Release.Info.LastDeployed.Time}),
	}

	return nil
}

func (c Chart) diagnose(s string) error {
	if s != "deployed" {
		return fmt.Errorf("chart is in an invalid state")
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// ReleaseRes represents an helm chart resource.
type ReleaseRes struct {
	Release *release.Release
}

// GetObjectKind returns a schema object.
func (ReleaseRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (h ReleaseRes) DeepCopyObject() runtime.Object {
	return h
}
