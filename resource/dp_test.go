package resource_test

import (
	"testing"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
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
	assert.False(t, l.HasXRay())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 6, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestDeploymentListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sDeployment(), nil)
	l := resource.NewDeploymentListWithArgs("blee", resource.NewDeploymentWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
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
	return `typemeta:
  kind: Deployment
  apiversion: apps/v1
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
spec:
  replicas: 1
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
  strategy:
    type: ""
    rollingupdate: null
  minreadyseconds: 0
  revisionhistorylimit: null
  paused: false
  progressdeadlineseconds: null
status:
  observedgeneration: 0
  replicas: 0
  updatedreplicas: 0
  readyreplicas: 0
  availablereplicas: 0
  unavailablereplicas: 0
  conditions: []
  collisioncount: null
`
}
