package watch

import (
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodList(t *testing.T) {
	cmo := NewMockConnection()
	no := NewPod(cmo, "")

	o := no.List("", metav1.ListOptions{})
	assert.Assert(t, o == nil)
}

func TestPodGet(t *testing.T) {
	cmo := NewMockConnection()
	no := NewPod(cmo, "")

	o, err := no.Get("", metav1.GetOptions{})
	assert.ErrorContains(t, err, "not found")
	assert.Assert(t, o == nil)
}
