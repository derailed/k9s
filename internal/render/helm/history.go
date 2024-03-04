// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package helm

import (
	"context"
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
)

// History renders a History chart to screen.
type History struct{}

// Healthy checks component health.
func (History) Healthy(ctx context.Context, o interface{}) error {
	return nil
}

// IsGeneric identifies a generic handler.
func (History) IsGeneric() bool {
	return false
}

// ColorerFunc colors a resource row.
func (History) ColorerFunc() model1.ColorerFunc {
	return model1.DefaultColorer
}

// Header returns a header row.
func (History) Header(_ string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "REVISION"},
		model1.HeaderColumn{Name: "STATUS"},
		model1.HeaderColumn{Name: "CHART"},
		model1.HeaderColumn{Name: "APP VERSION"},
		model1.HeaderColumn{Name: "DESCRIPTION"},
		model1.HeaderColumn{Name: "VALID", Wide: true},
	}
}

// Render renders a chart to screen.
func (c History) Render(o interface{}, ns string, r *model1.Row) error {
	h, ok := o.(ReleaseRes)
	if !ok {
		return fmt.Errorf("expected HistoryRes, but got %T", o)
	}

	r.ID = client.FQN(h.Release.Namespace, h.Release.Name)
	r.ID += ":" + strconv.Itoa(h.Release.Version)
	r.Fields = model1.Fields{
		strconv.Itoa(h.Release.Version),
		h.Release.Info.Status.String(),
		h.Release.Chart.Metadata.Name + "-" + h.Release.Chart.Metadata.Version,
		h.Release.Chart.Metadata.AppVersion,
		h.Release.Info.Description,
		render.AsStatus(c.diagnose(h.Release.Info.Status.String())),
	}

	return nil
}

func (c History) diagnose(s string) error {
	return nil
}
