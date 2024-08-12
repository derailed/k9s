// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package internal_test

import (
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/stretchr/testify/assert"
)

func TestIsLabelSelector(t *testing.T) {
	uu := map[string]struct {
		s  string
		ok bool
	}{
		"empty":       {s: ""},
		"cool":        {s: "-l app=fred,env=blee", ok: true},
		"no-flag":     {s: "app=fred,env=blee", ok: true},
		"no-space":    {s: "-lapp=fred,env=blee", ok: true},
		"wrong-flag":  {s: "-f app=fred,env=blee"},
		"missing-key": {s: "=fred"},
		"missing-val": {s: "fred="},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.ok, internal.IsLabelSelector(u.s))
		})
	}
}
