package views

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewStatus(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := newStatusView(defaults)
	v.update([]string{"blee", "duh"})

	assert.Equal(t, "[black:aqua:b] blee [-:black:-] [black:orange:b] duh [-:black:-] \n", v.GetText(false))
}
