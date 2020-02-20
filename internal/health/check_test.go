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
		c.Inc(health.OK)
		cc = append(cc, c)
		n++
	}
	c.Total(n)

	assert.Equal(t, 10, len(cc))
	assert.Equal(t, 10, c.Tally(health.Corpus))
	assert.Equal(t, 10, c.Tally(health.OK))
	assert.Equal(t, 0, c.Tally(health.Toast))
}
