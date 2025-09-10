// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_cleanseResource(t *testing.T) {
	uu := map[string]struct {
		r, e string
	}{
		"empty": {},
		"single": {
			r: "fred",
			e: "fred",
		},
		"grp/res": {
			r: "fred/blee",
			e: "blee",
		},
		"grp/res/sub": {
			r: "fred/blee/bob",
			e: "blee/bob",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, cleanseResource(u.r))
		})
	}
}
