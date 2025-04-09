// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	cfg "github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGenericRender(t *testing.T) {
	uu := map[string]struct {
		ns      string
		table   *metav1beta1.Table
		eID     string
		eFields model1.Fields
		eHeader model1.Header
	}{
		"withNS": {
			ns:      "ns1",
			table:   makeNSGeneric(),
			eID:     "ns1/fred",
			eFields: model1.Fields{"ns1", "c1", "c2", "c3"},
			eHeader: model1.Header{
				model1.HeaderColumn{Name: "NAMESPACE"},
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B"},
				model1.HeaderColumn{Name: "C"},
			},
		},

		"all": {
			ns:      client.NamespaceAll,
			table:   makeNSGeneric(),
			eID:     "ns1/fred",
			eFields: model1.Fields{"ns1", "c1", "c2", "c3"},
			eHeader: model1.Header{
				model1.HeaderColumn{Name: "NAMESPACE"},
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B"},
				model1.HeaderColumn{Name: "C"},
			},
		},

		"clusterWide": {
			ns:      client.ClusterScope,
			table:   makeNoNSGeneric(),
			eID:     "-/fred",
			eFields: model1.Fields{"c1", "c2", "c3"},
			eHeader: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B"},
				model1.HeaderColumn{Name: "C"},
			},
		},

		"age": {
			ns:      client.ClusterScope,
			table:   makeAgeGeneric(),
			eID:     "-/fred",
			eFields: model1.Fields{"c1", "c2", "2d"},
			eHeader: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "C"},
				model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
			},
		},
	}

	for k := range uu {
		var re render.Table
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			var r model1.Row
			re.SetTable(u.ns, u.table)

			assert.Equal(t, u.eHeader, re.Header(u.ns))
			require.NoError(t, re.Render(u.table.Rows[0], u.ns, &r))
			assert.Equal(t, u.eID, r.ID)
			assert.Equal(t, u.eFields, r.Fields)
		})
	}
}

func TestGenericCustRender(t *testing.T) {
	uu := map[string]struct {
		ns      string
		table   *metav1beta1.Table
		vs      cfg.ViewSetting
		eID     string
		eFields model1.Fields
		eHeader model1.Header
	}{
		"spec": {
			ns:    "ns1",
			table: makeNSGeneric(),
			vs: cfg.ViewSetting{
				Columns: []string{
					"NAMESPACE",
					"BLEE:.metadata.name",
					"ZORG:.metadata.namespace",
				},
			},
			eID:     "ns1/fred",
			eFields: model1.Fields{"ns1", "fred", "ns1", "c1", "c2", "c3"},
			eHeader: model1.Header{
				model1.HeaderColumn{Name: "NAMESPACE"},
				model1.HeaderColumn{Name: "BLEE"},
				model1.HeaderColumn{Name: "ZORG"},
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B"},
				model1.HeaderColumn{Name: "C"},
			},
		},
	}

	for k, u := range uu {
		var re render.Table
		re.SetViewSetting(&u.vs)
		t.Run(k, func(t *testing.T) {
			var r model1.Row
			re.SetTable(u.ns, u.table)

			assert.Equal(t, u.eHeader, re.Header(u.ns))
			require.NoError(t, re.Render(u.table.Rows[0], u.ns, &r))
			assert.Equal(t, u.eID, r.ID)
			assert.Equal(t, u.eFields, r.Fields)
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeNSGeneric() *metav1beta1.Table {
	return &metav1beta1.Table{
		ColumnDefinitions: []metav1beta1.TableColumnDefinition{
			{Name: "NAMESPACE"},
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		},
		Rows: []metav1beta1.TableRow{
			{
				Object: runtime.RawExtension{
					Object: &unstructured.Unstructured{
						Object: map[string]any{
							"kind":       "fred",
							"apiVersion": "v1",
							"metadata": map[string]any{
								"namespace": "ns1",
								"name":      "fred",
							},
						},
					},
				},
				Cells: []any{
					"ns1",
					"c1",
					"c2",
					"c3",
				},
			},
		},
	}
}

func makeNoNSGeneric() *metav1beta1.Table {
	return &metav1beta1.Table{
		ColumnDefinitions: []metav1beta1.TableColumnDefinition{
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		},
		Rows: []metav1beta1.TableRow{
			{
				Object: runtime.RawExtension{
					Object: &unstructured.Unstructured{
						Object: map[string]any{
							"kind":       "fred",
							"apiVersion": "v1",
							"metadata": map[string]any{
								"name": "fred",
							},
						},
					},
				},
				Cells: []any{
					"c1",
					"c2",
					"c3",
				},
			},
		},
	}
}

func makeAgeGeneric() *metav1beta1.Table {
	return &metav1beta1.Table{
		ColumnDefinitions: []metav1beta1.TableColumnDefinition{
			{Name: "a"},
			{Name: "Age"},
			{Name: "c"},
		},
		Rows: []metav1beta1.TableRow{
			{
				Object: runtime.RawExtension{
					Object: &unstructured.Unstructured{
						Object: map[string]any{
							"kind":       "fred",
							"apiVersion": "v1",
							"metadata": map[string]any{
								"name": "fred",
							},
						},
					},
				},
				Cells: []any{
					"c1",
					"2d",
					"c2",
				},
			},
		},
	}
}
