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

func TestDeploymentListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewDeploymentList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "deploy", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestDeploymentHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "DESIRED", "CURRENT", "UP-TO-DATE", "AVAILABLE", "AGE"}, newDeployment().Header(resource.DefaultNamespace))
}

func TestDeploymentFields(t *testing.T) {
	r := newDeployment().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestDeploymentMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sDeployment(), nil)

	cm := resource.NewDeploymentWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, dpYaml(), ma)
}

func TestDeploymentListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sDeployment()}, nil)

	l := resource.NewDeploymentListWithArgs("-", resource.NewDeploymentWithArgs(ca))
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
	assert.Equal(t, 6, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sDeployment() *appsv1.Deployment {
	var i int32 = 1
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &i,
		},
	}
}

func newDeployment() resource.Columnar {
	return resource.NewDeployment().NewInstance(k8sDeployment())
}

func dpYaml() string {
	return `apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  replicas: 1
  selector: null
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers: null
status: {}
`
}
