// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	uu := map[string]struct {
		s, e string
	}{
		"empty": {},
		"max": {
			s: "/app.kubernetes.io/instance=prom,app.kubernetes.io/name=prometheus,app.kubernetes.io/component=server",
			e: "/app.kubernetes.io/instance=prom,app.kubernetes.i…",
		},
		"less": {
			s: "app=fred,env=blee",
			e: "app=fred,env=blee",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, render.Truncate(u.s, 50))
		})
	}
}

func TestTrimLabelSelector(t *testing.T) {
	uu := map[string]struct {
		sel, e string
	}{
		"cool":    {"-l app=fred,env=blee", "app=fred,env=blee"},
		"noSpace": {"-lapp=fred,env=blee", "app=fred,env=blee"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, TrimLabelSelector(u.sel))
		})
	}
}
