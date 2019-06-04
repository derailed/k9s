package watch

import (
	"sync"
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	m "github.com/petergtz/pegomock"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestInformerInitWithNS(t *testing.T) {
	ns := "ns1"

	f := new(genericclioptions.ConfigFlags)
	f.Namespace = &ns
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	m.When(cmo.HasMetrics()).ThenReturn(true)
	m.When(cmo.CanIAccess("", "", "namespaces", []string{"list", "watch"})).ThenReturn(false, nil)
	m.When(cmo.CanIAccess("", ns, "namespaces", []string{"get", "watch"})).ThenReturn(true, nil)
	m.When(cmo.CanIAccess("", ns, "metrics.k8s.io", []string{"list", "watch"})).ThenReturn(true, nil)
	i := NewInformer(cmo, ns)

	o, err := i.List(PodIndex, "fred", metav1.ListOptions{})
	assert.NilError(t, err)
	assert.Assert(t, len(o) == 0)
}

func TestInformerList(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	i := NewInformer(cmo, "")

	o, err := i.List(PodIndex, "fred", metav1.ListOptions{})
	assert.NilError(t, err)
	assert.Assert(t, len(o) == 0)
}

func TestInformerListNoRes(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	i := NewInformer(cmo, "")

	o, err := i.List("dp", "fred", metav1.ListOptions{})
	assert.ErrorContains(t, err, "No informer found")
	assert.Assert(t, len(o) == 0)
}

func TestInformerGet(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	i := NewInformer(cmo, "")

	o, err := i.Get(PodIndex, "fred", metav1.GetOptions{})
	assert.ErrorContains(t, err, "Pod fred not found")
	assert.Assert(t, o == nil)
}

func TestInformerGetNoRes(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	i := NewInformer(cmo, "")

	o, err := i.Get("rs", "fred", metav1.GetOptions{})
	assert.ErrorContains(t, err, "No informer found")
	assert.Assert(t, o == nil)
}

func TestInformerRun(t *testing.T) {
	f := new(genericclioptions.ConfigFlags)
	cmo := NewMockConnection()
	m.When(cmo.Config()).ThenReturn(k8s.NewConfig(f))
	i := NewInformer(cmo, "")

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
