package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestIndicatorReset(t *testing.T) {
	s, _ := config.NewStyles("")

	i := ui.NewIndicatorView(ui.NewApp(), s)
	i.SetPermanent("Blee")
	i.Info("duh")
	i.Reset()

	assert.Equal(t, "Blee\n", i.GetText(false))
}

func TestIndicatorInfo(t *testing.T) {
	s, _ := config.NewStyles("")

	i := ui.NewIndicatorView(ui.NewApp(), s)
	i.Info("Blee")

	assert.Equal(t, "[lawngreen::b] <Blee> \n", i.GetText(false))
}

func TestIndicatorWarn(t *testing.T) {
	s, _ := config.NewStyles("")

	i := ui.NewIndicatorView(ui.NewApp(), s)
	i.Warn("Blee")

	assert.Equal(t, "[mediumvioletred::b] <Blee> \n", i.GetText(false))
}

func TestIndicatorErr(t *testing.T) {
	s, _ := config.NewStyles("")

	i := ui.NewIndicatorView(ui.NewApp(), s)
	i.Err("Blee")

	assert.Equal(t, "[orangered::b] <Blee> \n", i.GetText(false))
}
