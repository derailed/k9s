package resource_test

import (
	"testing"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestJobListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewJobList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "job", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestJobHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "COMPLETIONS", "DURATION", "AGE"}, newJob().Header(resource.DefaultNamespace))
}

func TestJobFields(t *testing.T) {
	r := newJob().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestJobMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sJob(), nil)

	cm := resource.NewJobWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, jobYaml(), ma)
}

func TestJobListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sJob()}, nil)

	l := resource.NewJobListWithArgs("-", resource.NewJobWithArgs(ca))
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
	assert.Equal(t, 4, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestJobListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sJob(), nil)
	l := resource.NewJobListWithArgs("blee", resource.NewJobWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sJob() *v1.Job {
	var i int32
	return &v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.JobSpec{
			Completions: &i,
			Parallelism: &i,
		},
		Status: v1.JobStatus{
			StartTime:      &metav1.Time{Time: testTime()},
			CompletionTime: &metav1.Time{Time: testTime()},
		},
	}
}

func newJob() resource.Columnar {
	return resource.NewJob().NewInstance(k8sJob())
}

func jobYaml() string {
	return `typemeta:
  kind: Job
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
  parallelism: 0
  completions: 0
  activedeadlineseconds: null
  backofflimit: null
  selector: null
  manualselector: null
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
  ttlsecondsafterfinished: null
status:
  conditions: []
  starttime: "2018-12-14T10:36:43.326972-07:00"
  completiontime: "2018-12-14T10:36:43.326972-07:00"
  active: 0
  succeeded: 0
  failed: 0
`
}
