package watch

import (
	"testing"

	"github.com/rs/zerolog"
	"gotest.tools/assert"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestPodMXList(t *testing.T) {
	cmo := NewMockConnection()
	no := NewPodMetrics(cmo, "")

	o := no.List("")
	assert.Assert(t, len(o) == 0)
}

func TestPodMXGet(t *testing.T) {
	cmo := NewMockConnection()
	no := NewPodMetrics(cmo, "")

	o, err := no.Get("")
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

	w.Run()
	w.Stop()
}
