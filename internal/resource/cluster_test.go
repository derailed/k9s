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
	setup(t)

	cIfc, mxIfc := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(cIfc.Version()).ThenReturn("1.2.3", nil)

	ci := resource.NewClusterWithArgs(cIfc, mxIfc)
	assert.Equal(t, "1.2.3", ci.Version())
}

func TestClusterNoVersion(t *testing.T) {
	setup(t)

	cIfc, mxIfc := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(cIfc.Version()).ThenReturn("bad", fmt.Errorf("No data"))

	ci := resource.NewClusterWithArgs(cIfc, mxIfc)
	assert.Equal(t, "n/a", ci.Version())
}

func TestClusterName(t *testing.T) {
	setup(t)

	cIfc, mxIfc := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(cIfc.ClusterName()).ThenReturn("fred")

	ci := resource.NewClusterWithArgs(cIfc, mxIfc)
	assert.Equal(t, "fred", ci.ClusterName())
}

func TestClusterMetrics(t *testing.T) {
	setup(t)

	cIfc, mxIfc := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mxIfc.ClusterLoad([]v1.Node{}, []mv1beta1.NodeMetrics{})).ThenReturn(clusterMetric())

	c := resource.NewClusterWithArgs(cIfc, mxIfc)
	assert.Equal(t, clusterMetric(), c.Metrics([]v1.Node{}, []mv1beta1.NodeMetrics{}))
}

// Helpers...

func setup(t *testing.T) {
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
