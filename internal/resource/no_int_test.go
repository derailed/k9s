package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
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

	no := NewNode(nil, nil)
	for _, u := range uu {
		cond := no.status(u.s, false)
		assert.Equal(t, "Ready", cond)
	}
}
