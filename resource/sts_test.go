package resource_test

import (
	"testing"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStsListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewStatefulSetList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, l.GetNamespace(), ns)
	assert.Equal(t, "sts", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestStsHeader(t *testing.T) {
	s := newSts()
	e := append(resource.Row{"NAMESPACE"}, stsHeader()...)
	assert.Equal(t, e, s.Header(resource.AllNamespaces))
	assert.Equal(t, stsHeader(), s.Header("fred"))
}

func TestStsFields(t *testing.T) {
	uu := []struct {
		i resource.Columnar
		e resource.Row
	}{
		{i: newSts(), e: resource.Row{"blee", "fred", "0", "1"}},
	}

	for _, u := range uu {
		assert.Equal(t, "blee/fred", u.i.Name())
		assert.Equal(t, u.e, u.i.Fields(resource.AllNamespaces)[:4])
		assert.Equal(t, u.e[1:4], u.i.Fields("blee")[:3])
	}
}

func TestSTSMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sSTS(), nil)

	cm := resource.NewStatefulSetWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, stsYaml(), ma)
}

func TestSTSListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sSTS()}, nil)

	l := resource.NewStatefulSetListWithArgs("blee", resource.NewStatefulSetWithArgs(ca))
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
	assert.Equal(t, 4, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestSTSListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sSTS(), nil)
	l := resource.NewStatefulSetListWithArgs("blee", resource.NewStatefulSetWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sSTS() *v1.StatefulSet {
	return &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			Namespace:         "blee",
			CreationTimestamp: metav1.Time{testTime()},
		},
		Spec: v1.StatefulSetSpec{
			Replicas: new(int32),
		},
		Status: v1.StatefulSetStatus{
			ReadyReplicas: 1,
		},
	}
}

func newSts() resource.Columnar {
	return resource.NewStatefulSet().NewInstance(k8sSTS())
}

func stsHeader() resource.Row {
	return resource.Row{"NAME", "DESIRED", "CURRENT", "AGE"}
}

func stsYaml() string {
	return `typemeta:
  kind: StatefulSet
  apiversion: v1
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
  replicas: 0
  selector: null
  template:
    objectmeta:
      name: ""
      generatename: ""
      namespace: ""
      selflink: ""
      uid: ""
      resourceversion: ""
      generation: 0
      creationtimestamp: "0001-01-01T00:00:00Z"
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
      volumes: []
      initcontainers: []
      containers: []
      restartpolicy: ""
      terminationgraceperiodseconds: null
      activedeadlineseconds: null
      dnspolicy: ""
      nodeselector: {}
      serviceaccountname: ""
      deprecatedserviceaccount: ""
      automountserviceaccounttoken: null
      nodename: ""
      hostnetwork: false
      hostpid: false
      hostipc: false
      shareprocessnamespace: null
      securitycontext: null
      imagepullsecrets: []
      hostname: ""
      subdomain: ""
      affinity: null
      schedulername: ""
      tolerations: []
      hostaliases: []
      priorityclassname: ""
      priority: null
      dnsconfig: null
      readinessgates: []
      runtimeclassname: null
      enableservicelinks: null
  volumeclaimtemplates: []
  servicename: ""
  podmanagementpolicy: ""
  updatestrategy:
    type: ""
    rollingupdate: null
  revisionhistorylimit: null
status:
  observedgeneration: 0
  replicas: 0
  readyreplicas: 1
  currentreplicas: 0
  updatedreplicas: 0
  currentrevision: ""
  updaterevision: ""
  collisioncount: null
  conditions: []
`
}
