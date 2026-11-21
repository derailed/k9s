// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToggledSuspend(t *testing.T) {
	yes, no := true, false

	uu := map[string]struct {
		suspend *bool
		want    bool
	}{
		"unset":     {want: true},
		"running":   {suspend: &no, want: true},
		"suspended": {suspend: &yes, want: false},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			got := toggledSuspend(u.suspend)
			assert.Equal(t, u.want, *got)
		})
	}
}
