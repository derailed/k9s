// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package helm

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
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
func (Chart) ColorerFunc() model1.ColorerFunc {
	return model1.DefaultColorer
}

// Header returns a header row.
func (Chart) Header(_ string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "REVISION"},
		model1.HeaderColumn{Name: "STATUS"},
		model1.HeaderColumn{Name: "CHART"},
		model1.HeaderColumn{Name: "APP VERSION"},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a chart to screen.
func (c Chart) Render(o interface{}, ns string, r *model1.Row) error {
	h, ok := o.(ReleaseRes)
	if !ok {
		return fmt.Errorf("expected ReleaseRes, but got %T", o)
	}

	r.ID = client.FQN(h.Release.Namespace, h.Release.Name)
	r.Fields = model1.Fields{
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

// ReleaseRes represents a helm chart resource.
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
