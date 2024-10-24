// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
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
func (e *Event) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		idx, ok := h.IndexOf("REASON", true)
		if ok && strings.TrimSpace(re.Row.Fields[idx]) == "Killing" {
			return model1.KillColor
		}

		return model1.DefaultColorer(ns, h, re)
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

func (e *Event) Header(ns string) model1.Header {
	if e.table == nil {
		return model1.Header{}
	}
	hh := make(model1.Header, 0, len(e.table.ColumnDefinitions))
	hh = append(hh, model1.HeaderColumn{Name: "NAMESPACE"})
	for _, h := range e.table.ColumnDefinitions {
		header := model1.HeaderColumn{Name: strings.ToUpper(h.Name)}
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
func (e *Event) Render(o interface{}, ns string, r *model1.Row) error {
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
	r.Fields = make(model1.Fields, 0, len(e.Header(ns)))
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
