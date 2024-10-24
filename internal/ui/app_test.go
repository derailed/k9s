// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestAppGetCmd(t *testing.T) {
	a := ui.NewApp(mock.NewMockConfig(), "")
	a.Init()
	a.CmdBuff().SetText("blee", "")

	assert.Equal(t, "blee", a.GetCmd())
}

func TestAppInCmdMode(t *testing.T) {
	a := ui.NewApp(mock.NewMockConfig(), "")
	a.Init()
	a.CmdBuff().SetText("blee", "")
	assert.False(t, a.InCmdMode())

	a.CmdBuff().SetActive(false)
	assert.False(t, a.InCmdMode())
}

func TestAppResetCmd(t *testing.T) {
	a := ui.NewApp(mock.NewMockConfig(), "")
	a.Init()
	a.CmdBuff().SetText("blee", "")

	a.ResetCmd()

	assert.Equal(t, "", a.CmdBuff().GetText())
}

func TestAppHasCmd(t *testing.T) {
	a := ui.NewApp(mock.NewMockConfig(), "")
	a.Init()

	a.ActivateCmd(true)
	assert.False(t, a.HasCmd())

	a.CmdBuff().SetText("blee", "")
	assert.True(t, a.InCmdMode())
}

func TestAppGetActions(t *testing.T) {
	a := ui.NewApp(mock.NewMockConfig(), "")
	a.Init()

	a.GetActions().Add(ui.KeyZ, ui.KeyAction{Description: "zorg"})

	assert.Equal(t, 6, a.GetActions().Len())
}

func TestAppViews(t *testing.T) {
	a := ui.NewApp(mock.NewMockConfig(), "")
	a.Init()

	vv := []string{"crumbs", "logo", "prompt", "menu"}
	for i := range vv {
		v := vv[i]
		t.Run(v, func(t *testing.T) {
			assert.NotNil(t, a.Views()[v])
		})
	}

	assert.NotNil(t, a.Crumbs())
	assert.NotNil(t, a.Logo())
	assert.NotNil(t, a.Prompt())
	assert.NotNil(t, a.Menu())
}
