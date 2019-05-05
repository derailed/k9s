package watch

import (
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "github.com/stretchr/testify/assert"
)

func TestContainerGet(t *testing.T) {
	cmo := NewMockConnection()

	c := NewContainer(NewPod(cmo, ""))

	o, err := c.Get("fred", metav1.GetOptions{})
	assert.ErrorContains(t, err, "not found")
	assert.Assert(t, o == nil)
}

func TestContainerList(t *testing.T) {
	cmo := NewMockConnection()

	c := NewContainer(NewPod(cmo, ""))

	o := c.List("fred", metav1.ListOptions{})
	assert.Assert(t, o == nil)
}

// ----------------------------------------------------------------------------
// Helpers...
