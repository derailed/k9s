package views

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdUpdate(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := newCmdView(defaults, 'T')
	v.update("blee")

	assert.Equal(t, "T> blee\n", v.GetText(false))
}

func TestNewCmdActivate(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := newCmdView(defaults, 'T')
	v.update("blee")
	v.append('!')

	assert.Equal(t, "T> blee!\n", v.GetText(false))
	assert.False(t, v.inCmdMode())
	v.active(true)
	assert.True(t, v.inCmdMode())
}
