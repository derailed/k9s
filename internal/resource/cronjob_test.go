package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
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
	return `apiVersion: extensions/batchv1beta1
kind: CronJob
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  jobTemplate:
    metadata:
      creationTimestamp: null
    spec:
      template:
        metadata:
          creationTimestamp: null
        spec:
          containers: null
  schedule: '*/1 * * * *'
  suspend: false
status:
  lastScheduleTime: "2018-12-14T17:36:43Z"
`
}
