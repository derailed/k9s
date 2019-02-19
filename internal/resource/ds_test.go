package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDSListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewDaemonSetList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "ds", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestDSHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "DESIRED", "CURRENT", "READY", "UP-TO-DATE", "AVAILABLE", "NODE_SELECTOR", "AGE"}, newDS().Header(resource.DefaultNamespace))
}

func TestDSFields(t *testing.T) {
	r := newDS().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestDSMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sDS(), nil)

	cm := resource.NewDaemonSetWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, dsYaml(), ma)
}

func TestDSListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sDS()}, nil)

	l := resource.NewDaemonSetListWithArgs("-", resource.NewDaemonSetWithArgs(ca))
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
	assert.Equal(t, 8, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestDSListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sDS(), nil)
	l := resource.NewDaemonSetListWithArgs("blee", resource.NewDaemonSetWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sDS() *extv1beta1.DaemonSet {
	return &extv1beta1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: extv1beta1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"fred": "blee"},
			},
		},
		Status: extv1beta1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			CurrentNumberScheduled: 1,
			NumberReady:            1,
			NumberAvailable:        1,
		},
	}
}

func newDS() resource.Columnar {
	return resource.NewDaemonSet().NewInstance(k8sDS())
}

func dsYaml() string {
	return `typemeta:
  kind: DaemonSet
  apiversion: extensions/v1beta1
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
  selector:
    matchlabels:
      fred: blee
    matchexpressions: []
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
  updatestrategy:
    type: ""
    rollingupdate: null
  minreadyseconds: 0
  templategeneration: 0
  revisionhistorylimit: null
status:
  currentnumberscheduled: 1
  numbermisscheduled: 0
  desirednumberscheduled: 1
  numberready: 1
  observedgeneration: 0
  updatednumberscheduled: 0
  numberavailable: 1
  numberunavailable: 0
  collisioncount: null
  conditions: []
`
}
