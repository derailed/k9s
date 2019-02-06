package resource_test

import (
	"testing"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	resv1 "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPVCListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewPVCList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "pvc", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestPVCHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "STATUS", "VOLUME", "CAPACITY", "ACCESS MODES", "STORAGECLASS", "AGE"}, newPVC().Header(resource.DefaultNamespace))
}

func TestPVCFields(t *testing.T) {
	r := newPVC().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestPVCMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sPVC(), nil)

	cm := resource.NewPVCWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, pvcYaml(), ma)
}

func TestPVCListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sPVC()}, nil)

	l := resource.NewPVCListWithArgs("-", resource.NewPVCWithArgs(ca))
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
	assert.Equal(t, 7, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestPVCListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sPVC(), nil)
	l := resource.NewPVCListWithArgs("blee", resource.NewPVCWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sPVC() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			VolumeName: "duh",
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resv1.Quantity{},
				},
			},
		},
	}
}

func newPVC() resource.Columnar {
	return resource.NewPVC().NewInstance(k8sPVC())
}

func pvcYaml() string {
	return `typemeta:
  kind: PersistentVolumeClaim
  apiversion: v1
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
  accessmodes: []
  selector: null
  resources:
    limits: {}
    requests:
      storage:
        format: ""
  volumename: duh
  storageclassname: null
  volumemode: null
  datasource: null
status:
  phase: ""
  accessmodes: []
  capacity: {}
  conditions: []
`
}
