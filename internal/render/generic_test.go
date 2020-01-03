package render_test

import (
	"testing"

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
		eHeader render.HeaderRow
	}{
		"specific_ns": {
			ns:      "blee",
			table:   makeNSGeneric(),
			eID:     "ns1/c1",
			eFields: render.Fields{"c1", "c2", "c3"},
			eHeader: render.HeaderRow{
				render.Header{Name: "A"},
				render.Header{Name: "B"},
				render.Header{Name: "C"},
			},
		},
		"all_ns": {
			ns:      "",
			table:   makeAllNSGeneric(),
			eID:     "ns1/c1",
			eFields: render.Fields{"ns1", "c1", "c2", "c3"},
			eHeader: render.HeaderRow{
				render.Header{Name: "NAMESPACE"},
				render.Header{Name: "A"},
				render.Header{Name: "B"},
				render.Header{Name: "C"},
			},
		},
		"cluster": {
			ns:      "-",
			table:   makeClusterGeneric(),
			eID:     "c1",
			eFields: render.Fields{"c1", "c2", "c3"},
			eHeader: render.HeaderRow{
				render.Header{Name: "A"},
				render.Header{Name: "B"},
				render.Header{Name: "C"},
			},
		},
		"age": {
			ns:      "-",
			table:   makeAgeGeneric(),
			eID:     "c1",
			eFields: render.Fields{"c1", "c2", "Age"},
			eHeader: render.HeaderRow{
				render.Header{Name: "A"},
				render.Header{Name: "C"},
				render.Header{Name: "AGE"},
			},
		},
	}

	var re render.Generic
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			var r render.Row
			re.SetTable(u.table)
			assert.Nil(t, re.Render(&u.table.Rows[0], u.ns, &r))
			assert.Equal(t, u.eID, r.ID)
			assert.Equal(t, u.eFields, r.Fields)
			assert.Equal(t, u.eHeader, re.Header(u.ns))
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

func makeAllNSGeneric() *metav1beta1.Table {
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

func makeClusterGeneric() *metav1beta1.Table {
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
					"Age",
					"c2",
				},
			},
		},
	}
}
