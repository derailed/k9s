package watch

import (
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeMXList(t *testing.T) {
	cmo := NewMockConnection()
	no := NewNodeMetrics(cmo)

	o := no.List("", metav1.ListOptions{})
	assert.Assert(t, len(o) == 0)
}

func TestNodeMXGet(t *testing.T) {
	cmo := NewMockConnection()
	no := NewNodeMetrics(cmo)

	o, err := no.Get("", metav1.GetOptions{})
	assert.ErrorContains(t, err, "No node metrics")
	assert.Assert(t, o == nil)
}

func TestNodeMXRun(t *testing.T) {
	cmo := NewMockConnection()
	w := newNodeMxWatcher(cmo)

	w.Run()
	w.Stop()
}
