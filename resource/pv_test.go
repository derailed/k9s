package resource_test

import (
	"testing"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPVListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewPVList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.Equal(t, "pv", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestPVHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "CAPACITY", "ACCESS MODES", "RECLAIM POLICY", "STATUS", "CLAIM", "STORAGECLASS", "REASON", "AGE"}, newPV().Header(resource.DefaultNamespace))
}

func TestPVFields(t *testing.T) {
	r := newPV().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestPVMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sPV(), nil)

	cm := resource.NewPVWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, pvYaml(), ma)
}

func TestPVListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sPV()}, nil)

	l := resource.NewPVListWithArgs("-", resource.NewPVWithArgs(ca))
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
	assert.Equal(t, 9, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestPVListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sPV(), nil)
	l := resource.NewPVListWithArgs("blee", resource.NewPVWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sPV() *v1.PersistentVolume {
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.PersistentVolumeSpec{},
	}
}

func newPV() resource.Columnar {
	return resource.NewPV().NewInstance(k8sPV())
}

func pvYaml() string {
	return `typemeta:
  kind: PV
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
  managedfields: []
spec:
  capacity: {}
  persistentvolumesource:
    gcepersistentdisk: null
    awselasticblockstore: null
    hostpath: null
    glusterfs: null
    nfs: null
    rbd: null
    iscsi: null
    cinder: null
    cephfs: null
    fc: null
    flocker: null
    flexvolume: null
    azurefile: null
    vspherevolume: null
    quobyte: null
    azuredisk: null
    photonpersistentdisk: null
    portworxvolume: null
    scaleio: null
    local: null
    storageos: null
    csi: null
  accessmodes: []
  claimref: null
  persistentvolumereclaimpolicy: ""
  storageclassname: ""
  mountoptions: []
  volumemode: null
  nodeaffinity: null
status:
  phase: ""
  message: ""
  reason: ""
`
}
