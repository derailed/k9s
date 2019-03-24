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

func TestEventListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewEventList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "event", l.GetName())
	for _, a := range []int{resource.ListAccess, resource.NamespaceAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestEventHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "REASON", "SOURCE", "COUNT", "MESSAGE", "AGE"}, newEvent().Header(resource.DefaultNamespace))
}

func TestEventFields(t *testing.T) {
	r := newEvent().Fields("blee")
	assert.Equal(t, resource.Row{"fred", "blah", "", "1"}, r[:4])
}

func TestEventMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sEvent(), nil)

	cm := resource.NewEventWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, evYaml(), ma)
}

func TestEventListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sEvent()}, nil)

	l := resource.NewEventListWithArgs("-", resource.NewEventWithArgs(ca))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced)
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
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
	return resource.NewEvent().NewInstance(k8sEvent())
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
