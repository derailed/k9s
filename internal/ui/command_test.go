package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestCmdNew(t *testing.T) {
	model := model.NewFishBuff(':', model.Command)
	v := ui.NewCommand(config.NewStyles(), model)

	model.AddListener(v)
	model.Set("blee")

	assert.Equal(t, "\x00> [::b]blee\n", v.GetText(false))
}

func TestCmdUpdate(t *testing.T) {
	model := model.NewFishBuff(':', model.Command)
	v := ui.NewCommand(config.NewStyles(), model)

	model.AddListener(v)
	model.Set("blee")
	model.Add('!')

	assert.Equal(t, "\x00> [::b]blee!\n", v.GetText(false))
	assert.False(t, v.InCmdMode())
}

func TestCmdMode(t *testing.T) {
	model := model.NewFishBuff(':', model.Command)
	v := ui.NewCommand(config.NewStyles(), model)
	model.AddListener(v)

	for _, f := range []bool{false, true} {
		model.SetActive(f)
		assert.Equal(t, f, v.InCmdMode())
	}
}
