// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tcell/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Event renders a K8s Event to screen.
type Event struct {
	Generic
}

func (*Event) IsGeneric() bool {
	return true
}

// ColorerFunc colors a resource row.
func (e *Event) ColorerFunc() ColorerFunc {
	return func(ns string, h Header, re RowEvent) tcell.Color {
		reasonCol := h.IndexOf("REASON", true)
		if reasonCol >= 0 && strings.TrimSpace(re.Row.Fields[reasonCol]) == "Killing" {
			return KillColor
		}

		return DefaultColorer(ns, h, re)
	}
}

var ageCols = map[string]struct{}{
	"FIRST SEEN": {},
	"LAST SEEN":  {},
}

var wideCols = map[string]struct{}{
	"SUBOBJECT":  {},
	"SOURCE":     {},
	"FIRST SEEN": {},
	"NAME":       {},
	"MESSAGE":    {},
}

func (e *Event) Header(ns string) Header {
	if e.table == nil {
		return Header{}
	}
	hh := make(Header, 0, len(e.table.ColumnDefinitions))
	hh = append(hh, HeaderColumn{Name: "NAMESPACE"})
	for _, h := range e.table.ColumnDefinitions {
		header := HeaderColumn{Name: strings.ToUpper(h.Name)}
		if _, ok := ageCols[header.Name]; ok {
			header.Time = true
		}
		if _, ok := wideCols[header.Name]; ok {
			header.Wide = true
		}
		hh = append(hh, header)
	}

	return hh
}

// Render renders a K8s resource to screen.
func (e *Event) Render(o interface{}, ns string, r *Row) error {
	row, ok := o.(metav1.TableRow)
	if !ok {
		return fmt.Errorf("expecting a TableRow but got %T", o)
	}
	nns, name, err := resourceNS(row.Object.Raw)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("expecting row 0 to be a string but got %T", row.Cells[0])
	}
	r.ID = client.FQN(nns, name)
	r.Fields = make(Fields, 0, len(e.Header(ns)))
	r.Fields = append(r.Fields, nns)
	for _, o := range row.Cells {
		if o == nil {
			r.Fields = append(r.Fields, Blank)
			continue
		}
		if s, ok := o.(fmt.Stringer); ok {
			r.Fields = append(r.Fields, s.String())
			continue
		}
		if s, ok := o.(string); ok {
			r.Fields = append(r.Fields, s)
			continue
		}
		r.Fields = append(r.Fields, fmt.Sprintf("%v", o))
	}

	return nil
}
