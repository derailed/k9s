package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

const ageTableCol = "Age"

// Generic renders a generic resource to screen.
type Generic struct {
	table *metav1beta1.Table

	ageIndex int
}

// Happy returns true if resoure is happy, false otherwise
func (Generic) Happy(ns string, r Row) bool {
	return true
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
func (g *Generic) Header(ns string) Header {
	if g.table == nil {
		return Header{}
	}
	h := make(Header, 0, len(g.table.ColumnDefinitions))
	h = append(h, HeaderColumn{Name: "NAMESPACE"})
	for i, c := range g.table.ColumnDefinitions {
		if c.Name == ageTableCol {
			g.ageIndex = i
			continue
		}
		h = append(h, HeaderColumn{Name: strings.ToUpper(c.Name)})
	}
	if g.ageIndex > 0 {
		h = append(h, HeaderColumn{Name: "AGE", Time: true})
	}

	return h
}

// Render renders a K8s resource to screen.
func (g *Generic) Render(o interface{}, ns string, r *Row) error {
	row, ok := o.(metav1beta1.TableRow)
	if !ok {
		return fmt.Errorf("expecting a TableRow but got %T", o)
	}
	nns, err := resourceNS(row.Object.Raw)
	if err != nil {
		return err
	}

	n, ok := row.Cells[0].(string)
	if !ok {
		return fmt.Errorf("expecting row 0 to be a string but got %T", row.Cells[0])
	}
	r.ID = client.FQN(nns, n)
	r.Fields = make(Fields, 0, len(g.Header(ns)))
	r.Fields = append(r.Fields, nns)
	var ageCell interface{}
	for i, c := range row.Cells {
		if g.ageIndex > 0 && i == g.ageIndex {
			ageCell = c
			continue
		}
		r.Fields = append(r.Fields, fmt.Sprintf("%v", c))
	}
	if ageCell != nil {
		r.Fields = append(r.Fields, fmt.Sprintf("%v", ageCell))
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func resourceNS(raw []byte) (string, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		return "", err
	}

	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return "", errors.New("no metadata found on generic resource")
	}

	ns, ok := meta["namespace"]
	if !ok {
		return client.ClusterScope, nil
	}

	nns, ok := ns.(string)
	if !ok {
		return "", fmt.Errorf("expecting namespace string type but got %T", ns)
	}
	return nns, nil
}
