package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestScrollIndicatorRefresg(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := view.NewAutoScrollIndicator(defaults)
	v.Refresh()

	assert.Equal(t, "[black:orange:b] Autoscroll: On  \n", v.GetText(false))
}
