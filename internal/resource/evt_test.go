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

func NewEventListWithArgs(ns string, r *resource.Event) resource.List {
	return resource.NewList(ns, "ev", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewEventWithArgs(conn k8s.Connection, res resource.Cruder) *resource.Event {
	r := &resource.Event{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestEventAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewEventListWithArgs(resource.AllNamespaces, NewEventWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "ev", l.GetName())
	for _, a := range []int{resource.ListAccess, resource.NamespaceAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestEventFields(t *testing.T) {
	r := newEvent().Fields("blee")
	assert.Equal(t, resource.Row{"fred", "blah", "", "1"}, r[:4])
}

func TestEventMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sEvent(), nil)

	cm := NewEventWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, evYaml(), ma)
}

func TestEventData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee")).ThenReturn(k8s.Collection{*k8sEvent()}, nil)

	l := NewEventListWithArgs("blee", NewEventWithArgs(mc, mr))
	// Make sure we mrn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
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

func k8sEvent() *v1.Event {
	return &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Reason:  "blah",
		Message: "blee",
		Count:   1,
	}
}

func newEvent() resource.Columnar {
	mc := NewMockConnection()
	return resource.NewEvent(mc).New(k8sEvent())
}

func evYaml() string {
	return `apiVersion: v1
count: 1
eventTime: null
firstTimestamp: null
involvedObject: {}
kind: Event
lastTimestamp: null
message: blee
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
reason: blah
reportingComponent: ""
reportingInstance: ""
source: {}
`
}
