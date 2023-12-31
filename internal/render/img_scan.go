// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/vul"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	CVEParseIdx = 5
	sevColName  = "SEVERITY"
)

// ImageScan renders scans report table.
type ImageScan struct {
	Base
}

// ColorerFunc colors a resource row.
func (c ImageScan) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		c := DefaultColorer(ns, h, re)

		sevCol := h.IndexOf(sevColName, true)
		if sevCol == -1 {
			return c
		}
		sev := strings.TrimSpace(re.Row.Fields[sevCol])
		switch sev {
		case vul.Sev1:
			c = tcell.ColorRed
		case vul.Sev2:
			c = tcell.ColorDarkOrange
		case vul.Sev3:
			c = tcell.ColorYellow
		case vul.Sev4:
			c = tcell.ColorDeepSkyBlue
		case vul.Sev5:
			c = tcell.ColorCadetBlue
		default:
			c = tcell.ColorDarkOliveGreen
		}

		return c
	}

}

// Header returns a header row.
func (ImageScan) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "SEVERITY"},
		HeaderColumn{Name: "VULNERABILITY"},
		HeaderColumn{Name: "IMAGE"},
		HeaderColumn{Name: "LIBRARY"},
		HeaderColumn{Name: "VERSION"},
		HeaderColumn{Name: "FIXED-IN"},
		HeaderColumn{Name: "TYPE"},
	}
}

// Render renders a K8s resource to screen.
func (is ImageScan) Render(o interface{}, name string, r *Row) error {
	res, ok := o.(ImageScanRes)
	if !ok {
		return fmt.Errorf("expected ImageScanRes, but got %T", o)
	}

	r.ID = fmt.Sprintf("%s|%s", res.Image, strings.Join(res.Row, "|"))
	r.Fields = Fields{
		res.Row.Severity(),
		res.Row.Vulnerability(),
		res.Image,
		res.Row.Name(),
		res.Row.Version(),
		res.Row.Fix(),
		res.Row.Type(),
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// ImageScanRes represents a container and its metrics.
type ImageScanRes struct {
	Image string
	Row   vul.Row
}

// GetObjectKind returns a schema object.
func (ImageScanRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (is ImageScanRes) DeepCopyObject() runtime.Object {
	return is
}
