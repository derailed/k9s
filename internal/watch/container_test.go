package watch

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
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

func TestToContainer(t *testing.T) {
	c := make(k8s.Collection, 2)
	toContainers(makeCoPod("p1"), c)

	assert.Equal(t, 2, len(c))
}

// ----------------------------------------------------------------------------
// Helpers...

func makePod(n string) *v1.Pod {
	po := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}
	po.Status.Phase = v1.PodRunning

	return po
}

func makeCoPod(n string) *v1.Pod {
	po := makePod(n)
	po.Spec = v1.PodSpec{
		InitContainers: []v1.Container{
			makeContainer("i1", "fred:0.0.1"),
		},
		Containers: []v1.Container{
			makeContainer("c1", "blee:0.1.0"),
		},
	}

	return po
}

func makeContainer(n, img string) v1.Container {
	co := v1.Container{
		Name:  n,
		Image: img,
	}

	return co
}
