package watch

import (
	"testing"

	"gotest.tools/assert"
	// "github.com/stretchr/testify/assert"
)

func TestContainerGet(t *testing.T) {
	cmo := NewMockConnection()

	c := NewContainer(NewPod(cmo, ""))

	o, err := c.Get("fred")
	assert.ErrorContains(t, err, "not found")
	assert.Assert(t, o == nil)
}

func TestContainerList(t *testing.T) {
	cmo := NewMockConnection()

	c := NewContainer(NewPod(cmo, ""))

	o := c.List("fred")
	assert.Assert(t, o == nil)
}

// ----------------------------------------------------------------------------
// Helpers...
