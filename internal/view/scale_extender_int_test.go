// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReplicas(t *testing.T) {
	uu := map[string]struct {
		s   string
		e   int32
		err bool
	}{
		"zero":     {s: "0", e: 0},
		"number":   {s: "3", e: 3},
		"leading":  {s: "01000", e: 1000},
		"empty":    {s: "", err: true},
		"dollar":   {s: "$1000", err: true},
		"space":    {s: " 1000", err: true},
		"negative": {s: "-1", err: true},
		"letters":  {s: "us", err: true},
		"overflow": {s: "4294967296", err: true},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			n, err := parseReplicas(u.s)
			if u.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, u.e, n)
		})
	}
}
