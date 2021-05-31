package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestLogIndicatorRefresh(t *testing.T) {
	defaults := config.NewStyles()
	uu := map[string]struct {
		li *view.LogIndicator
		e  string
	}{
		"all containers":    {view.NewLogIndicator(config.NewConfig(nil), defaults, true), "[::b]AllContainers:Off     [::b]Autoscroll:On     [::b]FullScreen:Off     [::b]Timestamps:Off     [::b]Wrap:Off\n"},
		"no all containers": {view.NewLogIndicator(config.NewConfig(nil), defaults, false), "[::b]Autoscroll:On     [::b]FullScreen:Off     [::b]Timestamps:Off     [::b]Wrap:Off\n"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.li.Refresh()
			assert.Equal(t, u.li.GetText(false), u.e)
		})
	}
}
