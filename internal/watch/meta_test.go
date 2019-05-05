package watch

import (
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMetaList(t *testing.T) {
	cmo := NewMockConnection()
	m := NewMeta(cmo, "")

	o, err := m.List(PodIndex, "fred", metav1.ListOptions{})
	assert.NilError(t, err)
	assert.Assert(t, len(o) == 0)
}

func TestMetaListNoRes(t *testing.T) {
	cmo := NewMockConnection()
	m := NewMeta(cmo, "")

	o, err := m.List("dp", "fred", metav1.ListOptions{})
	assert.ErrorContains(t, err, "No informer found")
	assert.Assert(t, len(o) == 0)
}

func TestMetaGet(t *testing.T) {
	cmo := NewMockConnection()
	m := NewMeta(cmo, "")

	o, err := m.Get(PodIndex, "fred", metav1.GetOptions{})
	assert.ErrorContains(t, err, "Pod fred not found")
	assert.Assert(t, o == nil)
}

func TestMetaGetNoRes(t *testing.T) {
	cmo := NewMockConnection()
	m := NewMeta(cmo, "")

	o, err := m.Get("rs", "fred", metav1.GetOptions{})
	assert.ErrorContains(t, err, "No informer found")
	assert.Assert(t, o == nil)
}

func TestMetaRun(t *testing.T) {
	cmo := NewMockConnection()
	m := NewMeta(cmo, "")

	c := make(chan struct{})
	m.Run(c)
	close(c)
}
