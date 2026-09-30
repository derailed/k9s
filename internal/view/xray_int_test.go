// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The filter text comes straight from the command buffer, so an unfinished
// regex like "[" is ordinary user input and must not take k9s down.
func TestXrayRxFilterInvalidRx(t *testing.T) {
	uu := map[string]struct {
		q       string
		e, eInv bool
	}{
		"valid":           {q: "pod", e: true, eInv: false},
		"unclosed class":  {q: "[", e: false, eInv: true},
		"dangling repeat": {q: "*", e: false, eInv: true},
		"unclosed group":  {q: "(", e: false, eInv: true},
		"invalid repeat":  {q: "a{2,1}", e: false, eInv: true},
		"unclosed escape": {q: `\`, e: false, eInv: true},
	}

	const path = "v1/pods::default/nginx"
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, rxFilter(u.q, path))
			assert.Equal(t, u.eInv, rxInverseFilter("!"+u.q, path))
		})
	}
}
