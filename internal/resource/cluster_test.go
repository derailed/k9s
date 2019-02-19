package resource_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

func TestClusterVersion(t *testing.T) {
	setup(t)

	cIfc, mxIfc := NewMockClusterIfc(), NewMockMetricsIfc()
	m.When(cIfc.Version()).ThenReturn("1.2.3", nil)

	ci := resource.NewClusterWithArgs(cIfc, mxIfc)
	assert.Equal(t, "1.2.3", ci.Version())
}

func TestClusterNoVersion(t *testing.T) {
	setup(t)

	cIfc, mxIfc := NewMockClusterIfc(), NewMockMetricsIfc()
	m.When(cIfc.Version()).ThenReturn("bad", fmt.Errorf("No data"))

	ci := resource.NewClusterWithArgs(cIfc, mxIfc)
	assert.Equal(t, "n/a", ci.Version())
}

func TestClusterName(t *testing.T) {
	setup(t)

	cIfc, mxIfc := NewMockClusterIfc(), NewMockMetricsIfc()
	m.When(cIfc.ClusterName()).ThenReturn("fred")

	ci := resource.NewClusterWithArgs(cIfc, mxIfc)
	assert.Equal(t, "fred", ci.ClusterName())
}

func TestClusterMetrics(t *testing.T) {
	setup(t)

	cIfc, mxIfc := NewMockClusterIfc(), NewMockMetricsIfc()
	m.When(mxIfc.NodeMetrics()).ThenReturn(testMetric(), nil)

	c := resource.NewClusterWithArgs(cIfc, mxIfc)
	m, err := c.Metrics()
	assert.Nil(t, err)
	assert.Equal(t, testMetric(), m)
}

// Helpers...

func setup(t *testing.T) {
	m.RegisterMockTestingT(t)
	m.RegisterMockFailHandler(func(m string, i ...int) {
		fmt.Println("Boom!", m, i)
	})
}

func testMetric() k8s.Metric {
	return k8s.Metric{
		CPU:      "100m",
		AvailCPU: "1000m",
		Mem:      "256Gi",
		AvailMem: "512Gi",
	}
}
