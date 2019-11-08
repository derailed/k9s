package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDaemonSetListWithArgs(ns string, r *resource.DaemonSet) resource.List {
	return resource.NewList(ns, "ds", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewDaemonSetWithArgs(conn k8s.Connection, res resource.Cruder) *resource.DaemonSet {
	r := &resource.DaemonSet{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestDSListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewDaemonSetListWithArgs(resource.AllNamespaces, NewDaemonSetWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "ds", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestDSFields(t *testing.T) {
	r := newDS().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestDSMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sDS(), nil)

	cm := NewDaemonSetWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, dsYaml(), ma)
}

func TestDSListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sDS()}, nil)

	l := NewDaemonSetListWithArgs("blee", NewDaemonSetWithArgs(mc, mr))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	mr.VerifyWasCalled(m.Times(2)).List("blee", metav1.ListOptions{})
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, "blee", l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 8, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sDS() *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"fred": "blee"},
			},
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			CurrentNumberScheduled: 1,
			NumberReady:            1,
			NumberAvailable:        1,
		},
	}
}

func newDS() resource.Columnar {
	mc := NewMockConnection()
	return resource.NewDaemonSet(mc).New(k8sDS())
}

func dsYaml() string {
	return `apiVersion: apps/v1
kind: DaemonSet
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  selector:
    matchLabels:
      fred: blee
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers: null
  updateStrategy: {}
status:
  currentNumberScheduled: 1
  desiredNumberScheduled: 1
  numberAvailable: 1
  numberMisscheduled: 0
  numberReady: 1
`
}
