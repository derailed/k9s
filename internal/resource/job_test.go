package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewJobListWithArgs(ns string, r *resource.Job) resource.List {
	return resource.NewList(ns, "job", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewJobWithArgs(conn k8s.Connection, res resource.Cruder) *resource.Job {
	r := &resource.Job{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestJobListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewJobListWithArgs(resource.AllNamespaces, NewJobWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "job", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestJobFields(t *testing.T) {
	r := newJob().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestJobMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sJob(), nil)

	cm := NewJobWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, jobYaml(), ma)
}

// BOZO!!
// func TestJobListData(t *testing.T) {
// 	mc := NewMockConnection()
// 	mr := NewMockCruder()
// 	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sJob()}, nil)

// 	l := NewJobListWithArgs("blee", NewJobWithArgs(mc, mr))
// 	// Make sure we mrn get deltas!
// 	for i := 0; i < 2; i++ {
// 		err := l.Reconcile(nil, "", "")
// 		assert.Nil(t, err)
// 	}

// 	mr.VerifyWasCalled(m.Times(2)).List("blee", metav1.ListOptions{})
// 	td := l.Data()
// 	assert.Equal(t, 1, len(td.Rows))
// 	assert.Equal(t, "blee", l.GetNamespace())
// 	row := td.Rows["blee/fred"]
// 	assert.Equal(t, 6, len(row.Deltas))
// 	for _, d := range row.Deltas {
// 		assert.Equal(t, "", d)
// 	}
// 	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
// }

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
	mc := NewMockConnection()
	c, _ := resource.NewJob(mc).New(k8sJob())
	return c
}

func jobYaml() string {
	return `apiVersion: extensions/v1beta1
kind: Job
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  completions: 0
  parallelism: 0
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers: null
status:
  completionTime: "2018-12-14T17:36:43Z"
  startTime: "2018-12-14T17:36:43Z"
`
}
