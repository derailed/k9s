package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGenericRender(t *testing.T) {
	uu := map[string]struct {
		ns      string
		table   *metav1beta1.Table
		eID     string
		eFields render.Fields
		eHeader render.Header
	}{
		"withNS": {
			ns:      "ns1",
			table:   makeNSGeneric(),
			eID:     "ns1/fred",
			eFields: render.Fields{"ns1", "c1", "c2", "c3"},
			eHeader: render.Header{
				render.HeaderColumn{Name: "NAMESPACE"},
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B"},
				render.HeaderColumn{Name: "C"},
			},
		},
		"all": {
			ns:      client.NamespaceAll,
			table:   makeNSGeneric(),
			eID:     "ns1/fred",
			eFields: render.Fields{"ns1", "c1", "c2", "c3"},
			eHeader: render.Header{
				render.HeaderColumn{Name: "NAMESPACE"},
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B"},
				render.HeaderColumn{Name: "C"},
			},
		},
		"allNS": {
			ns:      client.AllNamespaces,
			table:   makeNSGeneric(),
			eID:     "ns1/fred",
			eFields: render.Fields{"ns1", "c1", "c2", "c3"},
			eHeader: render.Header{
				render.HeaderColumn{Name: "NAMESPACE"},
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B"},
				render.HeaderColumn{Name: "C"},
			},
		},
		"clusterWide": {
			ns:      client.ClusterScope,
			table:   makeNoNSGeneric(),
			eID:     "-/fred",
			eFields: render.Fields{"-", "c1", "c2", "c3"},
			eHeader: render.Header{
				render.HeaderColumn{Name: "NAMESPACE"},
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B"},
				render.HeaderColumn{Name: "C"},
			},
		},
		"age": {
			ns:      client.ClusterScope,
			table:   makeAgeGeneric(),
			eID:     "-/fred",
			eFields: render.Fields{"-", "c1", "c2", "2d"},
			eHeader: render.Header{
				render.HeaderColumn{Name: "NAMESPACE"},
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "C"},
				render.HeaderColumn{Name: "AGE", Time: true},
			},
		},
	}

	for k := range uu {
		var re render.Generic
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			var r render.Row
			re.SetTable(u.table)

			assert.Equal(t, u.eHeader, re.Header(u.ns))
			assert.Nil(t, re.Render(u.table.Rows[0], u.ns, &r))
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
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		},
		Rows: []metav1beta1.TableRow{
			{
				Object: runtime.RawExtension{
					Raw: []byte(`{
        "kind": "fred",
        "apiVersion": "v1",
        "metadata": {
          "namespace": "ns1",
          "name": "fred"
        }}`),
				},
				Cells: []interface{}{
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
					Raw: []byte(`{
        "kind": "fred",
        "apiVersion": "v1",
        "metadata": {
          "name": "fred"
        }}`),
				},
				Cells: []interface{}{
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
					Raw: []byte(`{
        "kind": "fred",
        "apiVersion": "v1",
        "metadata": {
          "name": "fred"
        }}`),
				},
				Cells: []interface{}{
					"c1",
					"2d",
					"c2",
				},
			},
		},
	}
}
