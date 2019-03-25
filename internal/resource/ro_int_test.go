package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/rbac/v1"
)

func TestRoleParseRules(t *testing.T) {
	rules := []v1.PolicyRule{
		{
			Resources:       []string{"", "apps"},
			NonResourceURLs: []string{"/fred"},
			ResourceNames:   []string{"pods", "deployments"},
			Verbs:           []string{"get", "list"},
		},
	}

	var r Role
	rows := r.parseRules(rules)

	assert.Equal(t, 1, len(rows))
	assert.Equal(t, 1, len(rows))
}
