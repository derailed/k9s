// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
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
	sel, _ := labels.Parse("app=fred,env=blee")
	uu := map[string]struct {
		sel string
		err error
		e   labels.Selector
	}{
		"cool": {
			sel: "-l app=fred,env=blee",
			e:   sel,
		},

		"no-space": {
			sel: "-lapp=fred,env=blee",
			e:   sel,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			sel, err := TrimLabelSelector(u.sel)
			assert.Equal(t, u.err, err)
			assert.Equal(t, u.e, sel)
		})
	}
}
