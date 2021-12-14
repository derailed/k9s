package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell/v2"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
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
		if !Happy(ns, h, re.Row) {
			return ErrColor
		}
		reasonCol := h.IndexOf("REASON", true)
		if reasonCol == -1 {
			return DefaultColorer(ns, h, re)
		}
		if strings.TrimSpace(re.Row.Fields[reasonCol]) == "Killing" {
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
	row, ok := o.(metav1beta1.TableRow)
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
	for _, c := range row.Cells {
		if c == nil {
			r.Fields = append(r.Fields, Blank)
			continue
		}
		r.Fields = append(r.Fields, fmt.Sprintf("%v", c))
	}

	return nil
}
