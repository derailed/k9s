package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestAppGetCmd(t *testing.T) {
	a := ui.NewApp("")
	a.Init()
	a.CmdBuff().Set("blee")

	assert.Equal(t, "blee", a.GetCmd())
}

func TestAppInCmdMode(t *testing.T) {
	a := ui.NewApp("")
	a.Init()
	a.CmdBuff().Set("blee")
	assert.False(t, a.InCmdMode())

	a.CmdBuff().SetActive(true)
	assert.True(t, a.InCmdMode())
}

func TestAppResetCmd(t *testing.T) {
	a := ui.NewApp("")
	a.Init()
	a.CmdBuff().Set("blee")

	a.ResetCmd()

	assert.Equal(t, "", a.CmdBuff().String())
}

func TestAppHasCmd(t *testing.T) {
	a := ui.NewApp("")
	a.Init()

	a.ActivateCmd(true)
	assert.False(t, a.HasCmd())

	a.CmdBuff().Set("blee")
	assert.True(t, a.InCmdMode())
}

func TestAppGetActions(t *testing.T) {
	a := ui.NewApp("")
	a.Init()

	a.AddActions(ui.KeyActions{ui.KeyZ: ui.KeyAction{Description: "zorg"}})

	assert.Equal(t, 10, len(a.GetActions()))
}

func TestAppViews(t *testing.T) {
	a := ui.NewApp("")
	a.Init()

	vv := []string{"crumbs", "logo", "cmd", "menu"}
	for i := range vv {
		v := vv[i]
		t.Run(v, func(t *testing.T) {
			assert.NotNil(t, a.Views()[v])
		})
	}

	assert.NotNil(t, a.Crumbs())
	assert.NotNil(t, a.Logo())
	assert.NotNil(t, a.Cmd())
	assert.NotNil(t, a.Menu())
}
