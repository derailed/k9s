package watch

import (
	"testing"

	"gotest.tools/assert"
)

func TestMetaList(t *testing.T) {
	cmo := NewMockConnection()
	m := NewMeta(cmo, "")

	o, err := m.List(PodIndex, "fred")
	assert.NilError(t, err)
	assert.Assert(t, len(o) == 0)
}

func TestMetaListNoRes(t *testing.T) {
	cmo := NewMockConnection()
	m := NewMeta(cmo, "")

	o, err := m.List("dp", "fred")
	assert.ErrorContains(t, err, "No informer found")
	assert.Assert(t, len(o) == 0)
}

func TestMetaGet(t *testing.T) {
	cmo := NewMockConnection()
	m := NewMeta(cmo, "")

	o, err := m.Get(PodIndex, "fred")
	assert.ErrorContains(t, err, "Pod fred not found")
	assert.Assert(t, o == nil)
}

func TestMetaGetNoRes(t *testing.T) {
	cmo := NewMockConnection()
	m := NewMeta(cmo, "")

	o, err := m.Get("rs", "fred")
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
