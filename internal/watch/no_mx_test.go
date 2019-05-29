package watch

import (
	"sync"
	"testing"

	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
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

func TestNodeMXUpdate(t *testing.T) {
	cmo := NewMockConnection()
	no := newNodeMxWatcher(cmo)
	no.cache = map[string]runtime.Object{
		"n1": makeNodeMX("n1", "11m", "11Mi"),
	}

	mxx := &mv1beta1.NodeMetricsList{
		Items: []mv1beta1.NodeMetrics{
			*makeNodeMX("n1", "10m", "10Mi"),
		},
	}
	no.update(mxx, false)

	assert.Equal(t, toQty("10m"), *no.cache["n1"].(*mv1beta1.NodeMetrics).Usage.Cpu())
	assert.Equal(t, toQty("10Mi"), *no.cache["n1"].(*mv1beta1.NodeMetrics).Usage.Memory())
}

func TestNodeMXUpdateNoChange(t *testing.T) {
	cmo := NewMockConnection()
	no := newNodeMxWatcher(cmo)
	no.cache = map[string]runtime.Object{
		"n1": makeNodeMX("n1", "10m", "10Mi"),
	}

	mxx := &mv1beta1.NodeMetricsList{
		Items: []mv1beta1.NodeMetrics{
			*makeNodeMX("n1", "10m", "10Mi"),
		},
	}
	no.update(mxx, false)

	assert.Equal(t, toQty("10m"), *no.cache["n1"].(*mv1beta1.NodeMetrics).Usage.Cpu())
	assert.Equal(t, toQty("10Mi"), *no.cache["n1"].(*mv1beta1.NodeMetrics).Usage.Memory())
}

func TestNodeMXDelete(t *testing.T) {
	cmo := NewMockConnection()
	no := newNodeMxWatcher(cmo)
	no.cache = map[string]runtime.Object{
		"n1": makeNodeMX("n1", "11m", "11Mi"),
	}

	mxx := &mv1beta1.NodeMetricsList{}
	no.update(mxx, false)

	assert.Equal(t, 0, len(no.cache))
}

func TestNodeMXRun(t *testing.T) {
	cmo := NewMockConnection()
	w := newNodeMxWatcher(cmo)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Run()
	}()

	w.Stop()
	wg.Wait()
}

// ----------------------------------------------------------------------------
// Helpers...

func toQty(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)

	return q
}

func makeNodeMX(n, cpu, mem string) *v1beta1.NodeMetrics {
	return &v1beta1.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
		Usage: v1.ResourceList{
			v1.ResourceCPU:    toQty(cpu),
			v1.ResourceMemory: toQty(mem),
		},
	}
}
