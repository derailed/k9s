package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestNewLogoView(t *testing.T) {
	v := ui.NewLogo(config.NewStyles())
	v.Reset()

	const elogo = "[orange::b] ____  __.________       \n[orange::b]|    |/ _/   __   \\______\n[orange::b]|      < \\____    /  ___/\n[orange::b]|    |  \\   /    /\\___ \\ \n[orange::b]|____|__ \\ /____//____  >\n[orange::b]        \\/            \\/ \n"
	assert.Equal(t, elogo, v.Logo().GetText(false))
	assert.Equal(t, "", v.Status().GetText(false))
}

func TestLogoStatus(t *testing.T) {
	uu := map[string]struct {
		logo, msg, e string
	}{
		"info": {
			"[green::b] ____  __.________       \n[green::b]|    |/ _/   __   \\______\n[green::b]|      < \\____    /  ___/\n[green::b]|    |  \\   /    /\\___ \\ \n[green::b]|____|__ \\ /____//____  >\n[green::b]        \\/            \\/ \n",
			"blee",
			"[white::b]blee\n",
		},
		"warn": {
			"[mediumvioletred::b] ____  __.________       \n[mediumvioletred::b]|    |/ _/   __   \\______\n[mediumvioletred::b]|      < \\____    /  ___/\n[mediumvioletred::b]|    |  \\   /    /\\___ \\ \n[mediumvioletred::b]|____|__ \\ /____//____  >\n[mediumvioletred::b]        \\/            \\/ \n",
			"blee",
			"[white::b]blee\n",
		},
		"err": {
			"[red::b] ____  __.________       \n[red::b]|    |/ _/   __   \\______\n[red::b]|      < \\____    /  ___/\n[red::b]|    |  \\   /    /\\___ \\ \n[red::b]|____|__ \\ /____//____  >\n[red::b]        \\/            \\/ \n",
			"blee",
			"[white::b]blee\n",
		},
	}

	v := ui.NewLogo(config.NewStyles())
	for n := range uu {
		k, u := n, uu[n]
		t.Run(k, func(t *testing.T) {
			switch k {
			case "info":
				v.Info(u.msg)
			case "warn":
				v.Warn(u.msg)
			case "err":
				v.Err(u.msg)
			}
			assert.Equal(t, u.logo, v.Logo().GetText(false))
			assert.Equal(t, u.e, v.Status().GetText(false))
		})
	}

}
