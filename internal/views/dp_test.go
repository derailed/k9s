package views

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestDeployView(t *testing.T) {
	l := resource.NewDeploymentList(nil, "fred", k8s.GVR{})
	v := newDeployView("blee", NewApp(config.NewConfig(ks{})), l).(*deployView)

	assert.Equal(t, 3, len(v.hints()))
}
