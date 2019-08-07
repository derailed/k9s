package ui

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewLogoView(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := NewLogoView(defaults)
	v.Reset()

	const elogo = "[orange::b] ____  __.________       \n[orange::b]|    |/ _/   __   \\______\n[orange::b]|      < \\____    /  ___/\n[orange::b]|    |  \\   /    /\\___ \\ \n[orange::b]|____|__ \\ /____//____  >\n[orange::b]        \\/            \\/ \n"
	assert.Equal(t, elogo, v.logo.GetText(false))
	assert.Equal(t, "", v.status.GetText(false))
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

	defaults, _ := config.NewStyles("")
	v := NewLogoView(defaults)
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			switch k {
			case "info":
				v.Info(u.msg)
			case "warn":
				v.Warn(u.msg)
			case "err":
				v.Err(u.msg)
			}
			assert.Equal(t, u.logo, v.logo.GetText(false))
			assert.Equal(t, u.e, v.status.GetText(false))
		})
	}

}
