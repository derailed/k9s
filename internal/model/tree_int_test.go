// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The tree query is user typed, so an invalid regex must filter nothing out
// instead of panicking.
func TestRxMatchInvalidRx(t *testing.T) {
	uu := map[string]struct {
		q string
		e bool
	}{
		"valid":           {q: "nginx", e: true},
		"unclosed class":  {q: "[", e: false},
		"dangling repeat": {q: "*", e: false},
		"unclosed group":  {q: "(", e: false},
	}

	const path = "v1/pods::default/nginx"
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, rxMatch(u.q, path))
		})
	}
}
