package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGenericRender(t *testing.T) {
	var g render.Generic

	var r render.Row
	row := makeGeneric().Rows[0]
	assert.Nil(t, g.Render(&row, "blee", &r))

	assert.Equal(t, "blee/a", r.ID)
	assert.Equal(t, render.Fields{"a", "b", "c"}, r.Fields)
}

// Helpers...

func makeGeneric() *metav1beta1.Table {
	return &metav1beta1.Table{
		ColumnDefinitions: []metav1beta1.TableColumnDefinition{
			{Name: "A"},
			{Name: "B"},
			{Name: "C"},
		},
		Rows: []metav1beta1.TableRow{
			{
				Object: runtime.RawExtension{
					Raw: []byte(`{
        "kind": "fred",
        "apiVersion": "v1",
        "metadata": {
          "namespace": "blee",
          "name": "fred"
        }}`),
				},
				Cells: []interface{}{
					"a",
					"b",
					"c",
				},
			},
		},
	}
}
