package resource_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestClusterVersion(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.Version()).ThenReturn("1.2.3", nil)

	ci := resource.NewClusterWithArgs(mm, mx)
	assert.Equal(t, "1.2.3", ci.Version())
}

func TestClusterNoVersion(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.Version()).ThenReturn("bad", fmt.Errorf("No data"))

	ci := resource.NewClusterWithArgs(mm, mx)
	assert.Equal(t, "n/a", ci.Version())
}

func TestClusterName(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.ClusterName()).ThenReturn("fred")

	ci := resource.NewClusterWithArgs(mm, mx)
	assert.Equal(t, "fred", ci.ClusterName())
}

func TestContextName(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.ContextName()).ThenReturn("fred")

	ci := resource.NewClusterWithArgs(mm, mx)
	assert.Equal(t, "fred", ci.ContextName())
}

func TestUserName(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.UserName()).ThenReturn("fred")

	ci := resource.NewClusterWithArgs(mm, mx)
	assert.Equal(t, "fred", ci.UserName())
}

func TestClusterMetrics(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mx.ClusterLoad([]v1.Node{}, []mv1beta1.NodeMetrics{})).ThenReturn(clusterMetric())

	c := resource.NewClusterWithArgs(mm, mx)
	assert.Equal(t, clusterMetric(), c.Metrics([]v1.Node{}, []mv1beta1.NodeMetrics{}))
}

func TestClusterGetNodes(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.FetchNodes()).ThenReturn([]v1.Node{*k8sNode()}, nil)
	m.When(mx.ClusterLoad([]v1.Node{}, []mv1beta1.NodeMetrics{})).ThenReturn(clusterMetric())

	c := resource.NewClusterWithArgs(mm, mx)
	nodes, err := c.FetchNodes()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(nodes))
}

func TestClusterFetchNodesMetrics(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.FetchNodes()).ThenReturn([]v1.Node{*k8sNode()}, nil)
	m.When(mx.FetchNodesMetrics()).ThenReturn([]mv1beta1.NodeMetrics{makeMxNode("fred", "100m", "10Mi")}, nil)

	c := resource.NewClusterWithArgs(mm, mx)
	metrics, err := c.FetchNodesMetrics()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(metrics))
}

// Helpers...

func TestUsingMocks(t *testing.T) {
	m.RegisterMockTestingT(t)
	m.RegisterMockFailHandler(func(m string, i ...int) {
		fmt.Println("Boom!", m, i)
	})
}

func clusterMetric() k8s.ClusterMetrics {
	return k8s.ClusterMetrics{
		PercCPU: 100,
		PercMEM: 1000,
	}
}
