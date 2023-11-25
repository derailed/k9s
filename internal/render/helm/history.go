// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package helm

import (
	"context"
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
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
func (History) ColorerFunc() render.ColorerFunc {
	return render.DefaultColorer
}

// Header returns a header row.
func (History) Header(_ string) render.Header {
	return render.Header{
		render.HeaderColumn{Name: "REVISION"},
		render.HeaderColumn{Name: "STATUS"},
		render.HeaderColumn{Name: "CHART"},
		render.HeaderColumn{Name: "APP VERSION"},
		render.HeaderColumn{Name: "DESCRIPTION"},
		render.HeaderColumn{Name: "VALID", Wide: true},
	}
}

// Render renders a chart to screen.
func (c History) Render(o interface{}, ns string, r *render.Row) error {
	h, ok := o.(ReleaseRes)
	if !ok {
		return fmt.Errorf("expected HistoryRes, but got %T", o)
	}

	r.ID = client.FQN(h.Release.Namespace, h.Release.Name)
	r.ID += ":" + strconv.Itoa(h.Release.Version)
	r.Fields = render.Fields{
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
