package ui_test

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestFlashInfo(t *testing.T) {
	f := ui.NewFlash(ui.NewApp(""), "YO!")

	f.Info("Blee")
	assert.Equal(t, "😎 Blee\n", f.GetText(false))

	f.Infof("Blee %s", "duh")
	assert.Equal(t, "😎 Blee duh\n", f.GetText(false))
}

func TestFlashWarn(t *testing.T) {
	f := ui.NewFlash(ui.NewApp(""), "YO!")

	f.Warn("Blee")
	assert.Equal(t, "😗 Blee\n", f.GetText(false))

	f.Warnf("Blee %s", "duh")
	assert.Equal(t, "😗 Blee duh\n", f.GetText(false))
}

func TestFlashErr(t *testing.T) {
	f := ui.NewFlash(ui.NewApp(""), "YO!")

	f.Err(errors.New("Blee"))
	assert.Equal(t, "😡 Blee\n", f.GetText(false))

	f.Errf("Blee %s", "duh")
	assert.Equal(t, "😡 Blee duh\n", f.GetText(false))

}
