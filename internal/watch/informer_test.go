package watch

import (
	"errors"
	"sync"
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	m "github.com/petergtz/pegomock"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestInformerAllNSNoAccess(t *testing.T) {
	ns := ""
	f := new(genericclioptions.ConfigFlags)
	f.Namespace = &ns
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	m.When(cmo.HasMetrics()).ThenReturn(true)
	m.When(cmo.CheckListNSAccess()).ThenReturn(errors.New("denied"))
	m.When(cmo.CheckNSAccess(ns)).ThenReturn(errors.New("denied"))

	_, err := NewInformer(cmo, ns)
	assert.Error(t, err, "denied")
}

func TestInformerNSNoAccess(t *testing.T) {
	ns := "ns1"
	f := new(genericclioptions.ConfigFlags)
	f.Namespace = &ns
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	m.When(cmo.HasMetrics()).ThenReturn(true)
	m.When(cmo.CheckNSAccess(ns)).ThenReturn(errors.New("denied"))
	m.When(cmo.CheckListNSAccess()).ThenReturn(errors.New("denied"))

	_, err := NewInformer(cmo, ns)
	assert.Error(t, err, "denied")
}

func TestInformerInitWithNS(t *testing.T) {
	ns := "ns1"
	f := new(genericclioptions.ConfigFlags)
	f.Namespace = &ns
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	m.When(cmo.HasMetrics()).ThenReturn(true)
	m.When(cmo.CheckNSAccess(ns)).ThenReturn(nil)
	i, err := NewInformer(cmo, ns)
	assert.NilError(t, err)

	o, err := i.List(PodIndex, "fred", metav1.ListOptions{})
	assert.NilError(t, err)
	assert.Assert(t, len(o) == 0)
}

func TestInformerList(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	i, err := NewInformer(cmo, "")
	assert.NilError(t, err)

	o, err := i.List(PodIndex, "fred", metav1.ListOptions{})
	assert.NilError(t, err)
	assert.Assert(t, len(o) == 0)
}

func TestInformerListNoRes(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	i, err := NewInformer(cmo, "")
	assert.NilError(t, err)

	o, err := i.List("dp", "fred", metav1.ListOptions{})
	assert.ErrorContains(t, err, "No informer found")
	assert.Assert(t, len(o) == 0)
}

func TestInformerGet(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	i, err := NewInformer(cmo, "")
	assert.NilError(t, err)

	o, err := i.Get(PodIndex, "fred", metav1.GetOptions{})
	assert.ErrorContains(t, err, "Pod fred not found")
	assert.Assert(t, o == nil)
}

func TestInformerGetNoRes(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	i, err := NewInformer(cmo, "")
	assert.NilError(t, err)

	o, err := i.Get("rs", "fred", metav1.GetOptions{})
	assert.ErrorContains(t, err, "No informer found")
	assert.Assert(t, o == nil)
}

func TestInformerRun(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	i, err := NewInformer(cmo, "")
	assert.NilError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	c := make(chan struct{})
	go func() {
		defer wg.Done()
		i.Run(c)
	}()
	close(c)
	wg.Wait()
}
