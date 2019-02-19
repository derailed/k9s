package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReplicaSetMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sReplicaSet(), nil)

	cm := resource.NewReplicaSetWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, rsYaml(), ma)
}

func TestReplicaSetListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sReplicaSet()}, nil)

	l := resource.NewReplicaSetListWithArgs("blee", resource.NewReplicaSetWithArgs(ca))
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
	assert.Equal(t, 5, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestReplicaSetListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sReplicaSet(), nil)
	l := resource.NewReplicaSetListWithArgs("blee", resource.NewReplicaSetWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sReplicaSet() *v1.ReplicaSet {
	var i int32 = 1
	return &v1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.ReplicaSetSpec{
			Replicas: &i,
		},
		Status: v1.ReplicaSetStatus{
			ReadyReplicas: 1,
			Replicas:      1,
		},
	}
}

func newReplicaSet() resource.Columnar {
	return resource.NewReplicaSet().NewInstance(k8sReplicaSet())
}

func rsYaml() string {
	return `typemeta:
  kind: ReplicaSet
  apiversion: extensions/v1beta
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
  replicas: 1
  minreadyseconds: 0
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
status:
  replicas: 1
  fullylabeledreplicas: 0
  readyreplicas: 1
  availablereplicas: 0
  observedgeneration: 0
  conditions: []
`
}
