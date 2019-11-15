package watch

import (
	"strconv"
	"testing"

	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestMetaFQN(t *testing.T) {
	uu := map[string]struct {
		m metav1.ObjectMeta
		e string
	}{
		"full": {metav1.ObjectMeta{Namespace: "fred", Name: "blee"}, "fred/blee"},
		"nons": {metav1.ObjectMeta{Name: "blee"}, "blee"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, MetaFQN(u.m))
		})
	}
}

func TestMxResourceDiff(t *testing.T) {
	uu := map[string]struct {
		r1, r2 v1.ResourceList
		e      bool
	}{
		"same": {makeRes("0m", "0Mi"), makeRes("0m", "0Mi"), false},
		"omem": {makeRes("0m", "10Mi"), makeRes("0m", "1Mi"), true},
		"nmem": {makeRes("0m", "0Mi"), makeRes("0m", "1Mi"), true},
		"ocpu": {makeRes("1m", "0Mi"), makeRes("0m", "0Mi"), true},
		"ncpu": {makeRes("1m", "0Mi"), makeRes("2m", "0Mi"), true},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, resourceDiff(u.r1, u.r2))
		})
	}
}

func TestToSelector(t *testing.T) {
	uu := map[string]struct {
		s string
		e map[string]string
	}{
		"cool": {
			"app=fred,env=test",
			map[string]string{"app": "fred", "env": "test"},
		},
		"empty": {
			"",
			map[string]string{},
		},
		"hosed": {
			"app|blee",
			map[string]string{},
		},
		"toast": {
			"app,blee",
			map[string]string{},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			m := toSelector(u.s)
			for k, v := range m {
				assert.Equal(t, u.e[k], v)
			}
		})
	}
}

func TestMatchesNode(t *testing.T) {
	uu := map[string]struct {
		n string
		s map[string]string
		e bool
	}{
		"cool": {
			"n1",
			map[string]string{"spec.nodeName": "n1"},
			true,
		},
		"nomatch": {
			"n2",
			map[string]string{"spec.nodeName": "n1"},
			false,
		},
		"matchAll": {
			"n2",
			map[string]string{},
			true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, matchesNode(u.n, u.s))
		})
	}
}

func TestMatchesLabels(t *testing.T) {
	uu := map[string]struct {
		l, s map[string]string
		e    bool
	}{
		"cool": {
			map[string]string{"spec.nodeName": "n1"},
			map[string]string{"spec.nodeName": "n1"},
			true,
		},
		"nomatch": {
			map[string]string{"spec.nodeName": "n2"},
			map[string]string{"spec.nodeName": "n1"},
			false,
		},
		"matchAll": {
			map[string]string{"spec.nodeName": "n2"},
			map[string]string{},
			true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, matchesLabels(u.l, u.s))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeRes(c, m string) v1.ResourceList {
	cpu, _ := resource.ParseQuantity(c)
	mem, _ := resource.ParseQuantity(m)

	return v1.ResourceList{
		v1.ResourceCPU:    cpu,
		v1.ResourceMemory: mem,
	}
}

func makePodMxCo(name, cpu, mem string, co int) *mv1beta1.PodMetrics {
	mx := makePodMx(name)
	for i := 0; i < co; i++ {
		mx.Containers = append(
			mx.Containers,
			mv1beta1.ContainerMetrics{
				Name:  "c" + strconv.Itoa(i),
				Usage: makeRes(cpu, mem)})
	}

	return mx
}

func makePodMx(name string) *mv1beta1.PodMetrics {
	return &mv1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}
}
