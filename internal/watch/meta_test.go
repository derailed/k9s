package watch

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	m "github.com/petergtz/pegomock"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestMetaList(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	meta := NewMeta(cmo, "")

	o, err := meta.List(PodIndex, "fred", metav1.ListOptions{})
	assert.NilError(t, err)
	assert.Assert(t, len(o) == 0)
}

func TestMetaListNoRes(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	meta := NewMeta(cmo, "")

	o, err := meta.List("dp", "fred", metav1.ListOptions{})
	assert.ErrorContains(t, err, "No informer found")
	assert.Assert(t, len(o) == 0)
}

func TestMetaGet(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	meta := NewMeta(cmo, "")

	o, err := meta.Get(PodIndex, "fred", metav1.GetOptions{})
	assert.ErrorContains(t, err, "Pod fred not found")
	assert.Assert(t, o == nil)
}

func TestMetaGetNoRes(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	meta := NewMeta(cmo, "")

	o, err := meta.Get("rs", "fred", metav1.GetOptions{})
	assert.ErrorContains(t, err, "No informer found")
	assert.Assert(t, o == nil)
}

func TestMetaRun(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	meta := NewMeta(cmo, "")

	c := make(chan struct{})
	meta.Run(c)
	close(c)
}
