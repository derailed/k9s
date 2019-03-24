package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIngressListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewIngressList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "ing", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestIngressHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "HOSTS", "ADDRESS", "PORT", "AGE"}, newIngress().Header(resource.DefaultNamespace))
}

func TestIngressFields(t *testing.T) {
	r := newIngress().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestIngressMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sIngress(), nil)

	cm := resource.NewIngressWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, ingYaml(), ma)
}

func TestIngressListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sIngress()}, nil)

	l := resource.NewIngressListWithArgs("-", resource.NewIngressWithArgs(ca))
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
	assert.Equal(t, 5, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sIngress() *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{Host: "blee"},
			},
		},
	}
}

func newIngress() resource.Columnar {
	return resource.NewIngress().NewInstance(k8sIngress())
}

func ingYaml() string {
	return `apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  rules:
  - host: blee
status:
  loadBalancer: {}
`
}
