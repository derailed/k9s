package ui

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewCrumbs(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := NewCrumbsView(defaults)
	v.Refresh([]string{"blee", "duh"})

	assert.Equal(t, "[black:aqua:b] <blee> [-:black:-] [black:orange:b] <duh> [-:black:-] \n", v.GetText(false))
}
