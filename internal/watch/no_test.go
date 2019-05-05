package watch

import (
	"testing"

	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BenchmarkNodeFields(b *testing.B) {
	n := NewNode(nil)
	no := makeNode()
	ff := make(Row, 12)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		n.fields(no, ff)
	}
}

func TestJoin(t *testing.T) {
	uu := map[string]struct {
		i []string
		e string
	}{
		"zero":   {[]string{}, ""},
		"std":    {[]string{"a", "b", "c"}, "a,b,c"},
		"blank":  {[]string{"", "", ""}, ""},
		"sparse": {[]string{"a", "", "c"}, "a,c"},
	}

	for k, v := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, v.e, join(v.i, ","))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeNode() *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.NodeSpec{},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{Address: "1.1.1.1"},
			},
		},
	}
}
