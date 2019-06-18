package views

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestDaemonSetView(t *testing.T) {
	l := resource.NewDaemonSetList(nil, "fred")
	v := newDaemonSetView("blee", NewApp(config.NewConfig(ks{})), l)

	assert.Equal(t, "blee", v.getTitle())
}
