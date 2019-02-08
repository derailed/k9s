package resource_test

import (
	"testing"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCronJobListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewCronJobList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "job", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestCronJobHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "SCHEDULE", "SUSPEND", "ACTIVE", "LAST_SCHEDULE", "AGE"}, newCronJob().Header(resource.DefaultNamespace))
}

func TestCronJobFields(t *testing.T) {
	r := newCronJob().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestCronJobMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCronJob(), nil)

	cm := resource.NewCronJobWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, cronjobYaml(), ma)
}

func TestCronJobListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sCronJob()}, nil)

	l := resource.NewCronJobListWithArgs("-", resource.NewCronJobWithArgs(ca))
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

func TestCronJobListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCronJob(), nil)
	l := resource.NewCronJobListWithArgs("blee", resource.NewCronJobWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sCronJob() *batchv1beta1.CronJob {
	var b bool
	return &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule: "*/1 * * * *",
			Suspend:  &b,
		},
		Status: batchv1beta1.CronJobStatus{
			LastScheduleTime: &metav1.Time{Time: testTime()},
		},
	}
}

func newCronJob() resource.Columnar {
	return resource.NewCronJob().NewInstance(k8sCronJob())
}

func cronjobYaml() string {
	return `typemeta:
  kind: CronJob
  apiversion: extensions/batchv1beta1
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
  schedule: '*/1 * * * *'
  startingdeadlineseconds: null
  concurrencypolicy: ""
  suspend: false
  jobtemplate:
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
      parallelism: null
      completions: null
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
  successfuljobshistorylimit: null
  failedjobshistorylimit: null
status:
  active: []
  lastscheduletime: "2018-12-14T10:36:43.326972-07:00"
`
}
