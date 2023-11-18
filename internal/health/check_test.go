// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package health_test

import (
	"testing"

	"github.com/derailed/k9s/internal/health"
	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	var cc health.Checks

	c := health.NewCheck("test")
	n := 0
	for i := 0; i < 10; i++ {
		c.Inc(health.S1)
		cc = append(cc, c)
		n++
	}
	c.Total(int64(n))

	assert.Equal(t, 10, len(cc))
	assert.Equal(t, int64(10), c.Tally(health.Corpus))
	assert.Equal(t, int64(10), c.Tally(health.S1))
	assert.Equal(t, int64(0), c.Tally(health.S2))
}
