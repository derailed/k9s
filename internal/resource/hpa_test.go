package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	res "k8s.io/apimachinery/pkg/api/resource"

	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHPAListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewHPAList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "hpa", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestHPAHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "REFERENCE", "TARGETS", "MINPODS", "MAXPODS", "REPLICAS", "AGE"}, newHPA().Header(resource.DefaultNamespace))
}

func TestHPAFields(t *testing.T) {
	r := newHPA().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestHPAMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sHPA(), nil)

	cm := resource.NewHPAWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, hpaYaml(), ma)
}

func TestHPAListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sHPA()}, nil)

	l := resource.NewHPAListWithArgs("-", resource.NewHPAWithArgs(ca))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced)
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 7, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sHPA() *autoscalingv2beta2.HorizontalPodAutoscaler {
	var i int32 = 1
	return &autoscalingv2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
				Kind: "fred",
				Name: "blee",
			},
			MinReplicas: &i,
			MaxReplicas: 1,
			Metrics: []autoscalingv2beta2.MetricSpec{
				{
					Type: autoscalingv2beta2.ResourceMetricSourceType,
					Resource: &autoscalingv2beta2.ResourceMetricSource{
						Name: v1.ResourceCPU,
						Target: autoscalingv2beta2.MetricTarget{
							Type: autoscalingv2beta2.UtilizationMetricType,
						},
					},
				},
			},
		},
		Status: autoscalingv2beta2.HorizontalPodAutoscalerStatus{
			CurrentReplicas: 1,
			CurrentMetrics: []autoscalingv2beta2.MetricStatus{
				{
					Type: autoscalingv2beta2.ResourceMetricSourceType,
					Resource: &autoscalingv2beta2.ResourceMetricStatus{
						Name: v1.ResourceCPU,
						Current: autoscalingv2beta2.MetricValueStatus{
							Value: &res.Quantity{},
						},
					},
				},
			},
		},
	}
}

func newHPA() resource.Columnar {
	return resource.NewHPA().NewInstance(k8sHPA())
}

func hpaYaml() string {
	return `apiVersion: autoscaling/v2beta2
kind: HorizontalPodAutoscaler
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  maxReplicas: 1
  metrics:
  - resource:
      name: cpu
      target:
        type: Utilization
    type: Resource
  minReplicas: 1
  scaleTargetRef:
    kind: fred
    name: blee
status:
  conditions: null
  currentMetrics:
  - resource:
      current:
        value: "0"
      name: cpu
    type: Resource
  currentReplicas: 1
  desiredReplicas: 0
`
}
