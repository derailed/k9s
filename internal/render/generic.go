package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

// Generic renders a generic resource to screen.
type Generic struct {
	table *metav1beta1.Table
}

// SetTable sets the tabular resource.
func (g *Generic) SetTable(t *metav1beta1.Table) {
	g.table = t
}

// ColorerFunc colors a resource row.
func (Generic) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (g *Generic) Header(ns string) HeaderRow {
	if g.table == nil {
		return HeaderRow{}
	}

	h := make(HeaderRow, 0, len(g.table.ColumnDefinitions))
	if ns == "" {
		h = append(h, Header{Name: "NAMESPACE"})
	}
	for _, c := range g.table.ColumnDefinitions {
		h = append(h, Header{Name: strings.ToUpper(c.Name)})
	}

	return h
}

// Render renders a K8s resource to screen.
func (g *Generic) Render(o interface{}, ns string, r *Row) error {
	row, ok := o.(*metav1beta1.TableRow)
	if !ok {
		return fmt.Errorf("expecting a TableRow but got %T", o)
	}

	r.ID, ok = row.Cells[0].(string)
	if !ok {
		return fmt.Errorf("expecting row id to be a string but got %#v", row.Cells[0])
	}

	rns, err := extractNamespace(row.Object.Raw)
	if err != nil {
		return err
	}

	r.ID = FQN(rns, r.ID)
	r.Fields = make(Fields, 0, len(g.Header(ns)))
	if isAllNamespace(ns) {
		r.Fields = append(r.Fields, rns)
	}
	for _, c := range row.Cells {
		r.Fields = append(r.Fields, fmt.Sprintf("%v", c))
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func extractNamespace(raw []byte) (string, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		return "", err
	}

	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return "", errors.New("no metadata found on generic resource")
	}
	ns, ok := meta["namespace"].(string)
	if !ok {
		return "", errors.New("invalid namespace found on generic metadata")
	}

	return ns, nil
}
