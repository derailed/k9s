package render

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_defaultHeader(t *testing.T) {
	uu := map[string]struct {
		cdefs []metav1.TableColumnDefinition
		e     model1.Header
	}{
		"empty": {
			e: make(model1.Header, 0),
		},

		"plain": {
			cdefs: []metav1.TableColumnDefinition{
				{Name: "A"},
				{Name: "B"},
				{Name: "C"},
			},
			e: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B"},
				model1.HeaderColumn{Name: "C"},
			},
		},

		"age": {
			cdefs: []metav1.TableColumnDefinition{
				{Name: "Fred"},
				{Name: "Blee"},
				{Name: "Age"},
			},
			e: model1.Header{
				model1.HeaderColumn{Name: "FRED"},
				model1.HeaderColumn{Name: "BLEE"},
				model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
			},
		},

		"time-cols": {
			cdefs: []metav1.TableColumnDefinition{
				{Name: "Last Seen"},
				{Name: "Fred"},
				{Name: "Blee"},
				{Name: "Age"},
				{Name: "First Seen"},
			},
			e: model1.Header{
				model1.HeaderColumn{Name: "LAST SEEN", Attrs: model1.Attrs{Time: true}},
				model1.HeaderColumn{Name: "FRED"},
				model1.HeaderColumn{Name: "BLEE"},
				model1.HeaderColumn{Name: "FIRST SEEN", Attrs: model1.Attrs{Time: true}},
				model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			var ta Table
			ta.SetTable("ns-1", &metav1.Table{ColumnDefinitions: u.cdefs})
			assert.Equal(t, u.e, ta.defaultHeader())
		})
	}
}
