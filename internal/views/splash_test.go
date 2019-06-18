package views

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewSplash(t *testing.T) {
	defaults, _ := config.NewStyles("")
	s := newSplash(defaults, "bozo")

	x, y, w, h := s.GetRect()
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)
	assert.Equal(t, 15, w)
	assert.Equal(t, 10, h)
}
