package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestCmdNew(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := ui.NewCmdView(defaults)

	buff := ui.NewCmdBuff(':', ui.CommandBuff)
	buff.AddListener(v)
	buff.Set("blee")

	assert.Equal(t, "\x00> blee\n", v.GetText(false))
}

func TestCmdUpdate(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := ui.NewCmdView(defaults)

	buff := ui.NewCmdBuff(':', ui.CommandBuff)
	buff.AddListener(v)

	buff.Set("blee")
	buff.Add('!')

	assert.Equal(t, "\x00> blee!\n", v.GetText(false))
	assert.False(t, v.InCmdMode())
}

func TestCmdMode(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := ui.NewCmdView(defaults)

	buff := ui.NewCmdBuff(':', ui.CommandBuff)
	buff.AddListener(v)

	for _, f := range []bool{false, true} {
		buff.SetActive(f)
		assert.Equal(t, f, v.InCmdMode())
	}
}
