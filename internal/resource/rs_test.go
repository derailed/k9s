package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
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
	row := td.Rows["blee/fred"]
	assert.Equal(t, 5, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
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
	return `apiVersion: extensions/v1beta
kind: ReplicaSet
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  replicas: 1
  selector: null
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers: null
status:
  readyReplicas: 1
  replicas: 1
`
}
