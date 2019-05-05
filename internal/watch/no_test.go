package watch

import (
	"testing"

	"gotest.tools/assert"
)

func TestNodeList(t *testing.T) {
	cmo := NewMockConnection()
	no := NewNode(cmo)

	o := no.List("")
	assert.Assert(t, o == nil)
}

func TestNodeGet(t *testing.T) {
	cmo := NewMockConnection()
	no := NewNode(cmo)

	o, err := no.Get("")
	assert.ErrorContains(t, err, "not found")
	assert.Assert(t, o == nil)
}
