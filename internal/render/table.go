// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ageTableCol = "Age"

// Table renders a tabular resource to screen.
type Table struct {
	Base
	table    *metav1.Table
	header   model1.Header
	ageIndex int
}

func (*Table) IsGeneric() bool {
	return true
}

// SetTable sets the tabular resource.
func (t *Table) SetTable(ns string, table *metav1.Table) {
	t.table = table
	t.header = t.Header(ns)
}

// ColorerFunc colors a resource row.
func (*Table) ColorerFunc() model1.ColorerFunc {
	return model1.DefaultColorer
}

// Header returns a header row.
func (t *Table) Header(ns string) model1.Header {
	return t.doHeader(t.defaultHeader())
}

// Header returns a header row.
func (t *Table) defaultHeader() model1.Header {
	if t.table == nil {
		return model1.Header{}
	}
	h := make(model1.Header, 0, len(t.table.ColumnDefinitions))
	for i, c := range t.table.ColumnDefinitions {
		if c.Name == ageTableCol {
			t.ageIndex = i
			continue
		}
		h = append(h, model1.HeaderColumn{Name: strings.ToUpper(c.Name)})
	}
	if t.ageIndex > 0 {
		h = append(h, model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}})
	}

	return h
}

// Render renders a K8s resource to screen.
func (t *Table) Render(o any, ns string, r *model1.Row) error {
	row, ok := o.(metav1.TableRow)
	if !ok {
		return fmt.Errorf("expected Table, but got %T", o)
	}

	if err := t.defaultRow(&row, ns, r); err != nil {
		return err
	}
	if t.specs.isEmpty() {
		return nil
	}

	cols, err := t.specs.realize(row.Object.Object, t.defaultHeader(), r)
	if err != nil {
		return err
	}
	cols.hydrateRow(r)

	return nil
}

func (t *Table) defaultRow(row *metav1.TableRow, ns string, r *model1.Row) error {
	th := t.defaultHeader()
	ons, name := ns, UnknownValue
	switch {
	case row.Object.Object != nil:
		if m, _ := meta.Accessor(row.Object.Object); m != nil {
			ons, name = m.GetNamespace(), m.GetName()
		}
	case row.Object.Raw != nil:
		var pm metav1.PartialObjectMetadata
		if err := json.Unmarshal(row.Object.Raw, &pm); err != nil {
			return err
		}
		ons, name = pm.Namespace, pm.Name
	default:
		if idx, ok := th.IndexOf("NAME", true); ok && idx >= 0 {
			name = row.Cells[idx].(string)
			if idx, ok := th.IndexOf("NAMESPACE", true); ok && idx >= 0 {
				ons = row.Cells[idx].(string)
			}
		}
	}

	if client.IsClusterWide(ons) {
		ons = client.ClusterScope
	}
	r.ID = client.FQN(ons, name)
	r.Fields = make(model1.Fields, 0, len(th))
	var age any
	for i, c := range row.Cells {
		if t.ageIndex > 0 && i == t.ageIndex {
			age = c
			continue
		}
		if c == nil {
			r.Fields = append(r.Fields, Blank)
			continue
		}
		r.Fields = append(r.Fields, fmt.Sprintf("%v", c))
	}
	if d, ok := age.(string); ok {
		r.Fields = append(r.Fields, d)
	} else if t.ageIndex > 0 {
		slog.Warn("No Duration detected on age field")
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
