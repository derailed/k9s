package views

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestDaemonSetView(t *testing.T) {
	l := resource.NewDaemonSetList(nil, "fred", k8s.GVR{})
	v := newDaemonSetView("blee", NewApp(config.NewConfig(ks{})), l).(*daemonSetView)

	assert.Equal(t, 3, len(v.hints()))
}
