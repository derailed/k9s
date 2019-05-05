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

func NewServiceListWithArgs(ns string, r *resource.Service) resource.List {
	return resource.NewList(ns, "svc", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewServiceWithArgs(conn k8s.Connection, res resource.Cruder) *resource.Service {
	r := &resource.Service{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestSvcListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewServiceListWithArgs(resource.AllNamespaces, NewServiceWithArgs(mc, mr))
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
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sSVC(), nil)

	cm := NewServiceWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, svcYaml(), ma)
}

func TestSVCListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee")).ThenReturn(k8s.Collection{*k8sSVC()}, nil)

	l := NewServiceListWithArgs("blee", NewServiceWithArgs(mc, mr))
	// Make sure we mrn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	mr.VerifyWasCalled(m.Times(2)).List("blee")
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, "blee", l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 6, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
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
	mc := NewMockConnection()
	return resource.NewService(mc).New(k8sSVC())
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
