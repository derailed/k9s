package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestLogIndicatorRefresh(t *testing.T) {
	defaults := config.NewStyles()
	v := view.NewLogIndicator(defaults)
	v.Refresh()

	assert.Equal(t, "[black:orange:b] Autoscroll: On  [black:orange:b] FullScreen: Off [black:orange:b] Wrap: Off       \n", v.GetText(false))
}
