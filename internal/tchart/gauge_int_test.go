// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tchart

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeDeltas(t *testing.T) {
	uu := map[string]struct {
		d1, d2 int64
		e      delta
	}{
		"same": {
			e: DeltaSame,
		},
		"more": {
			d1: 10,
			d2: 20,
			e:  DeltaMore,
		},
		"less": {
			d1: 20,
			d2: 10,
			e:  DeltaLess,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, computeDelta(u.d1, u.d2))
		})
	}
}
