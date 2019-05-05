package watch

import (
	"testing"

	"gotest.tools/assert"
)

func TestNodeMXList(t *testing.T) {
	cmo := NewMockConnection()
	no := NewNodeMetrics(cmo)

	o := no.List("")
	assert.Assert(t, len(o) == 0)
}

func TestNodeMXGet(t *testing.T) {
	cmo := NewMockConnection()
	no := NewNodeMetrics(cmo)

	o, err := no.Get("")
	assert.ErrorContains(t, err, "No node metrics")
	assert.Assert(t, o == nil)
}

func TestNodeMXRun(t *testing.T) {
	cmo := NewMockConnection()
	w := newNodeMxWatcher(cmo)

	w.Run()
	w.Stop()
}
