// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package health_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/health"
	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	var cc health.Checks

	c := health.NewCheck(client.NewGVR("test"))
	n := 0
	for range 10 {
		c.Inc(health.S1)
		cc = append(cc, c)
		n++
	}
	c.Total(int64(n))

	assert.Len(t, cc, 10)
	assert.Equal(t, int64(10), c.Tally(health.Corpus))
	assert.Equal(t, int64(10), c.Tally(health.S1))
	assert.Equal(t, int64(0), c.Tally(health.S2))
}
