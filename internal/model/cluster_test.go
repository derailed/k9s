package model_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model"
	m "github.com/petergtz/pegomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestClusterVersion(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.Version()).ThenReturn("1.2.3", nil)

	ci := model.NewClusterWithArgs(mm, mx)
	assert.Equal(t, "1.2.3", ci.Version())
}

func TestClusterNoVersion(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.Version()).ThenReturn("bad", fmt.Errorf("No data"))

	ci := model.NewClusterWithArgs(mm, mx)
	assert.Equal(t, "n/a", ci.Version())
}

func TestClusterName(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.ClusterName()).ThenReturn("fred", nil)

	ci := model.NewClusterWithArgs(mm, mx)
	assert.Equal(t, "fred", ci.ClusterName())
}

func TestContextName(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.ContextName()).ThenReturn("fred", nil)

	ci := model.NewClusterWithArgs(mm, mx)
	assert.Equal(t, "fred", ci.ContextName())
}

func TestUserName(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()
	m.When(mm.UserName()).ThenReturn("fred", nil)

	ci := model.NewClusterWithArgs(mm, mx)
	assert.Equal(t, "fred", ci.UserName())
}

func TestClusterMetrics(t *testing.T) {
	mm, mx := NewMockClusterMeta(), NewMockMetricsServer()

	mxx := clusterMetric()

	c := model.NewClusterWithArgs(mm, mx)
	c.Metrics(nil, nil, &mxx)
	assert.Equal(t, clusterMetric(), mxx)
}

// Helpers...

func TestUsingMocks(t *testing.T) {
	m.RegisterMockTestingT(t)
	m.RegisterMockFailHandler(func(m string, i ...int) {
		fmt.Println("Boom!", m, i)
	})
}

func clusterMetric() client.ClusterMetrics {
	return client.ClusterMetrics{
		PercCPU: 100,
		PercMEM: 1000,
	}
}
