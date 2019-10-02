package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeStatus(t *testing.T) {
	uu := []struct {
		s v1.NodeStatus
		e string
	}{
		{
			v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
				},
			},
			"Ready",
		},
	}

	no := NewNode(nil, k8s.GVR{})
	for _, u := range uu {
		res := make([]string, 5)
		no.status(u.s, false, res)
		assert.Equal(t, "Ready", join(res, ","))
	}
}

func BenchmarkNodeFields(b *testing.B) {
	n := NewNode(nil, k8s.GVR{})
	no := makeNode()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = n.New(no).Fields("")
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
