package watch

import (
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestPodMXList(t *testing.T) {
	cmo := NewMockConnection()
	po := NewPodMetrics(cmo, "")

	o := po.List("", metav1.ListOptions{})
	assert.Assert(t, len(o) == 0)
}

func TestPodMXGet(t *testing.T) {
	cmo := NewMockConnection()
	po := NewPodMetrics(cmo, "")

	o, err := po.Get("", metav1.GetOptions{})
	assert.ErrorContains(t, err, "No pod metrics")
	assert.Assert(t, o == nil)
}

func TestMxDeltas(t *testing.T) {
	uu := map[string]struct {
		m1, m2 *mv1beta1.PodMetrics
		e      bool
	}{
		"same": {makePodMxCo("p1", "1m", "0Mi", 1), makePodMxCo("p1", "1m", "0Mi", 1), false},
		"dcpu": {makePodMxCo("p1", "10m", "0Mi", 1), makePodMxCo("p1", "0m", "0Mi", 1), true},
		"dmem": {makePodMxCo("p1", "0m", "10Mi", 1), makePodMxCo("p1", "0m", "0Mi", 1), true},
		"dco":  {makePodMxCo("p1", "0m", "10Mi", 1), makePodMxCo("p1", "0m", "0Mi", 2), true},
	}

	var p podMxWatcher
	for k, v := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, v.e, p.deltas(v.m1, v.m2))
		})
	}
}

func TestPodMXRun(t *testing.T) {
	cmo := NewMockConnection()
	w := newPodMxWatcher(cmo, "")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Run()
	}()

	w.Stop()
	wg.Wait()
}

func TestPodMXUpdate(t *testing.T) {
	cmo := NewMockConnection()
	po := newPodMxWatcher(cmo, "default")
	po.cache = map[string]runtime.Object{
		"default/p1": makePodMX("p1", "11m", "11Mi"),
	}

	mxx := &mv1beta1.PodMetricsList{
		Items: []mv1beta1.PodMetrics{
			*makePodMX("p1", "10m", "10Mi"),
		},
	}
	po.update(mxx, false)

	pmx := po.cache["default/p1"].(*mv1beta1.PodMetrics)
	assert.Equal(t, toQty("10m"), *pmx.Containers[0].Usage.Cpu())
	assert.Equal(t, toQty("10Mi"), *pmx.Containers[0].Usage.Memory())
}

func TestPodMXUpdateNoChange(t *testing.T) {
	cmo := NewMockConnection()
	po := newPodMxWatcher(cmo, "default")
	po.cache = map[string]runtime.Object{
		"default/p1": makePodMX("p1", "10m", "10Mi"),
	}

	mxx := &mv1beta1.PodMetricsList{
		Items: []mv1beta1.PodMetrics{
			*makePodMX("p1", "10m", "10Mi"),
		},
	}
	po.update(mxx, false)

	pmx := po.cache["default/p1"].(*mv1beta1.PodMetrics)
	assert.Equal(t, toQty("10m"), *pmx.Containers[0].Usage.Cpu())
	assert.Equal(t, toQty("10Mi"), *pmx.Containers[0].Usage.Memory())
}

func TestPodMXDelete(t *testing.T) {
	cmo := NewMockConnection()
	po := newPodMxWatcher(cmo, "default")
	po.cache = map[string]runtime.Object{
		"default/p1": makePodMX("p1", "11m", "11Mi"),
	}

	mxx := &mv1beta1.PodMetricsList{}
	po.update(mxx, false)

	assert.Equal(t, 0, len(po.cache))
}

// ----------------------------------------------------------------------------
// Helpers...

func makePodMX(name, cpu, mem string) *v1beta1.PodMetrics {
	return &v1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Containers: []v1beta1.ContainerMetrics{
			{Name: "i1", Usage: makeRes(cpu, mem)},
			{Name: "c1", Usage: makeRes(cpu, mem)},
		},
	}
}
