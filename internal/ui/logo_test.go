// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

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

	const elogo = "[#ffa500::b] ____  __.________       \n[#ffa500::b]|    |/ _/   __   \\______\n[#ffa500::b]|      < \\____    /  ___/\n[#ffa500::b]|    |  \\   /    /\\___ \\ \n[#ffa500::b]|____|__ \\ /____//____  >\n[#ffa500::b]        \\/            \\/ \n"
	assert.Equal(t, elogo, v.Logo().GetText(false))
	assert.Equal(t, "", v.Status().GetText(false))
}

func TestLogoStatus(t *testing.T) {
	uu := map[string]struct {
		logo, msg, e string
	}{
		"info": {
			"[#008000::b] ____  __.________       \n[#008000::b]|    |/ _/   __   \\______\n[#008000::b]|      < \\____    /  ___/\n[#008000::b]|    |  \\   /    /\\___ \\ \n[#008000::b]|____|__ \\ /____//____  >\n[#008000::b]        \\/            \\/ \n",
			"blee",
			"[#ffffff::b]blee\n",
		},
		"warn": {
			"[#c71585::b] ____  __.________       \n[#c71585::b]|    |/ _/   __   \\______\n[#c71585::b]|      < \\____    /  ___/\n[#c71585::b]|    |  \\   /    /\\___ \\ \n[#c71585::b]|____|__ \\ /____//____  >\n[#c71585::b]        \\/            \\/ \n",
			"blee",
			"[#ffffff::b]blee\n",
		},
		"err": {
			"[#ff0000::b] ____  __.________       \n[#ff0000::b]|    |/ _/   __   \\______\n[#ff0000::b]|      < \\____    /  ___/\n[#ff0000::b]|    |  \\   /    /\\___ \\ \n[#ff0000::b]|____|__ \\ /____//____  >\n[#ff0000::b]        \\/            \\/ \n",
			"blee",
			"[#ffffff::b]blee\n",
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
