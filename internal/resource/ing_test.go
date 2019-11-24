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

func NewIngressListWithArgs(ns string, r *resource.Ingress) resource.List {
	return resource.NewList(ns, "ing", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewIngressWithArgs(conn k8s.Connection, res resource.Cruder) *resource.Ingress {
	r := &resource.Ingress{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestIngressListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewIngressListWithArgs(resource.AllNamespaces, NewIngressWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "ing", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestIngressFields(t *testing.T) {
	r := newIngress().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestIngressMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sIngress(), nil)

	cm := NewIngressWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, ingYaml(), ma)
}

func TestIngressListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sIngress()}, nil)

	l := NewIngressListWithArgs("blee", NewIngressWithArgs(mc, mr))
	// Make sure we mrn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	mr.VerifyWasCalled(m.Times(2)).List("blee", metav1.ListOptions{})
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, "blee", l.GetNamespace())
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
	mc := NewMockConnection()
	c, _ := resource.NewIngress(mc).New(k8sIngress())
	return c
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
