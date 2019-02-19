package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSvcListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewServiceList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, l.GetNamespace(), ns)
	assert.Equal(t, "svc", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestSvcHeader(t *testing.T) {
	s := newSvc()
	e := append(resource.Row{"NAMESPACE"}, svcHeader()...)

	assert.Equal(t, e, s.Header(resource.AllNamespaces))
	assert.Equal(t, svcHeader(), s.Header("fred"))
}

func TestSvcFields(t *testing.T) {
	uu := []struct {
		i resource.Columnar
		e resource.Row
	}{
		{
			i: newSvc(),
			e: resource.Row{
				"blee",
				"fred",
				"ClusterIP",
				"1.1.1.1",
				"2.2.2.2",
				"http:90►0",
			},
		},
	}

	for _, u := range uu {
		assert.Equal(t, "blee/fred", u.i.Name())
		assert.Equal(t, u.e[1:6], u.i.Fields("blee")[:5])
		assert.Equal(t, u.e[:6], u.i.Fields(resource.AllNamespaces)[:6])
	}
}

func TestSVCMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sSVC(), nil)

	cm := resource.NewServiceWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, svcYaml(), ma)
}

func TestSVCListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sSVC()}, nil)

	l := resource.NewServiceListWithArgs("blee", resource.NewServiceWithArgs(ca))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List("blee")
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, "blee", l.GetNamespace())
	assert.False(t, l.HasXRay())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 6, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestSVCListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sSVC(), nil)
	l := resource.NewServiceListWithArgs("blee", resource.NewServiceWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sSVC() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			Namespace:         "blee",
			CreationTimestamp: metav1.Time{testTime()},
		},
		Spec: v1.ServiceSpec{
			Type:        v1.ServiceTypeClusterIP,
			ClusterIP:   "1.1.1.1",
			ExternalIPs: []string{"2.2.2.2"},
			Selector:    map[string]string{"fred": "blee"},
			Ports: []v1.ServicePort{
				{
					Name:     "http",
					Port:     90,
					Protocol: "TCP",
				},
			},
		},
	}
}

func newSvc() resource.Columnar {
	return resource.NewService().NewInstance(k8sSVC())
}

func svcHeader() resource.Row {
	return resource.Row{
		"NAME",
		"TYPE",
		"CLUSTER-IP",
		"EXTERNAL-IP",
		"PORT(S)",
		"AGE",
	}
}

func svcYaml() string {
	return `apiVersion: v1
kind: Service
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  clusterIP: 1.1.1.1
  externalIPs:
  - 2.2.2.2
  ports:
  - name: http
    port: 90
    protocol: TCP
    targetPort: 0
  selector:
    fred: blee
  type: ClusterIP
status:
  loadBalancer: {}
`
}
