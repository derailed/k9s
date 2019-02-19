package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/autoscaling/v1"
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
	assert.False(t, l.HasXRay())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 7, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestHPAListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sHPA(), nil)
	l := resource.NewHPAListWithArgs("blee", resource.NewHPAWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sHPA() *v1.HorizontalPodAutoscaler {
	var i int32 = 1
	return &v1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v1.CrossVersionObjectReference{
				Kind: "fred",
				Name: "blee",
			},
			MinReplicas:                    &i,
			MaxReplicas:                    1,
			TargetCPUUtilizationPercentage: &i,
		},
		Status: v1.HorizontalPodAutoscalerStatus{
			CurrentReplicas:                 1,
			CurrentCPUUtilizationPercentage: &i,
		},
	}
}

func newHPA() resource.Columnar {
	return resource.NewHPA().NewInstance(k8sHPA())
}

func hpaYaml() string {
	return `typemeta:
  kind: HPA
  apiversion: autoscaling/v1
objectmeta:
  name: fred
  generatename: ""
  namespace: blee
  selflink: ""
  uid: ""
  resourceversion: ""
  generation: 0
  creationtimestamp: "2018-12-14T10:36:43.326972-07:00"
  deletiontimestamp: null
  deletiongraceperiodseconds: null
  labels: {}
  annotations: {}
  ownerreferences: []
  initializers: null
  finalizers: []
  clustername: ""
  managedfields: []
spec:
  scaletargetref:
    kind: fred
    name: blee
    apiversion: ""
  minreplicas: 1
  maxreplicas: 1
  targetcpuutilizationpercentage: 1
status:
  observedgeneration: null
  lastscaletime: null
  currentreplicas: 1
  desiredreplicas: 0
  currentcpuutilizationpercentage: 1
`
}
