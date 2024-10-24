// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ageTableCol = "Age"

// Generic renders a generic resource to screen.
type Generic struct {
	Base
	table    *metav1.Table
	header   model1.Header
	ageIndex int
}

func (*Generic) IsGeneric() bool {
	return true
}

// SetTable sets the tabular resource.
func (g *Generic) SetTable(ns string, t *metav1.Table) {
	g.table = t
	g.header = g.Header(ns)
}

// ColorerFunc colors a resource row.
func (*Generic) ColorerFunc() model1.ColorerFunc {
	return model1.DefaultColorer
}

// Header returns a header row.
func (g *Generic) Header(ns string) model1.Header {
	if g.header != nil {
		return g.header
	}
	if g.table == nil {
		return model1.Header{}
	}
	h := make(model1.Header, 0, len(g.table.ColumnDefinitions))
	if !client.IsClusterScoped(ns) {
		h = append(h, model1.HeaderColumn{Name: "NAMESPACE"})
	}
	for i, c := range g.table.ColumnDefinitions {
		if c.Name == ageTableCol {
			g.ageIndex = i
			continue
		}
		h = append(h, model1.HeaderColumn{Name: strings.ToUpper(c.Name)})
	}
	if g.ageIndex > 0 {
		h = append(h, model1.HeaderColumn{Name: "AGE", Time: true})
	}

	return h
}

// Render renders a K8s resource to screen.
func (g *Generic) Render(o interface{}, ns string, r *model1.Row) error {
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
	r.Fields = make(model1.Fields, 0, len(g.Header(ns)))
	if !client.IsClusterScoped(ns) {
		r.Fields = append(r.Fields, nns)
	}
	var duration interface{}
	for i, c := range row.Cells {
		if g.ageIndex > 0 && i == g.ageIndex {
			duration = c
			continue
		}
		if c == nil {
			r.Fields = append(r.Fields, Blank)
			continue
		}
		r.Fields = append(r.Fields, fmt.Sprintf("%v", c))
	}
	if d, ok := duration.(string); ok {
		r.Fields = append(r.Fields, d)
	} else if g.ageIndex > 0 {
		log.Warn().Msgf("No Duration detected on age field")
		r.Fields = append(r.Fields, NAValue)
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func resourceNS(raw []byte) (string, string, error) {
	var obj map[string]interface{}
	var ns, name string
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		return ns, name, err
	}

	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return ns, name, errors.New("no metadata found on generic resource")
	}
	ina, ok := meta["name"]
	if !ok {
		return ns, name, errors.New("unable to extract resource name")
	}
	name, ok = ina.(string)
	if !ok {
		return ns, name, fmt.Errorf("expecting name string type but got %T", ns)
	}

	ins, ok := meta["namespace"]
	if !ok {
		return client.ClusterScope, name, nil
	}

	ns, ok = ins.(string)
	if !ok {
		return ns, name, fmt.Errorf("expecting namespace string type but got %T", ns)
	}
	return ns, name, nil
}
