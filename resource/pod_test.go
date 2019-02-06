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

func TestPodListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewPodList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "po", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestPodHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "READY", "STATUS", "RESTARTS", "CPU", "MEM", "IP", "NODE", "QOS", "AGE"}, newPod().Header(resource.DefaultNamespace))
}

func TestPodFields(t *testing.T) {
	r := newPod().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestPodMarshal(t *testing.T) {
	setup(t)

	mx := NewMockMetricsIfc()
	m.When(mx.PodMetrics()).ThenReturn(map[string]k8s.Metric{"fred": {}}, nil)
	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sPod(), nil)

	cm := resource.NewPodWithArgs(ca, mx)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, poYaml(), ma)
}

func TestPodListData(t *testing.T) {
	setup(t)

	mx := NewMockMetricsIfc()
	m.When(mx.PodMetrics()).ThenReturn(map[string]k8s.Metric{"fred": {}}, nil)
	ca := NewMockCaller()
	m.When(ca.List("")).ThenReturn(k8s.Collection{*k8sPod()}, nil)

	l := resource.NewPodListWithArgs("", resource.NewPodWithArgs(ca, mx))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List(resource.AllNamespaces)
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.AllNamespaces, l.GetNamespace())
	assert.True(t, l.HasXRay())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 11, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"blee"}, row.Fields[:1])
}

func TestPodListDescribe(t *testing.T) {
	setup(t)

	mx := NewMockMetricsIfc()
	m.When(mx.PodMetrics()).ThenReturn(map[string]k8s.Metric{"fred": {}}, nil)
	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sPod(), nil)
	l := resource.NewPodListWithArgs("blee", resource.NewPodWithArgs(ca, mx))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 8, len(props))
}

// Helpers...

func k8sPod() *v1.Pod {
	var i int32 = 1
	var t = v1.HostPathDirectory
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			Labels:            map[string]string{"blee": "duh"},
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.PodSpec{
			Priority:          &i,
			PriorityClassName: "bozo",
			Containers: []v1.Container{
				{
					Name:  "fred",
					Image: "blee",
					Env: []v1.EnvVar{
						{
							Name:  "fred",
							Value: "1",
							ValueFrom: &v1.EnvVarSource{
								ConfigMapKeyRef: &v1.ConfigMapKeySelector{Key: "blee"},
							},
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "fred",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/blee",
							Type: &t,
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: "Running",
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:         "fred",
					State:        v1.ContainerState{Running: &v1.ContainerStateRunning{}},
					RestartCount: 0,
				},
			},
		},
	}
}

func newPod() resource.Columnar {
	return resource.NewPod().NewInstance(k8sPod())
}

func poYaml() string {
	return `typemeta:
  kind: Pod
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
  labels:
    blee: duh
  annotations: {}
  ownerreferences: []
  initializers: null
  finalizers: []
  clustername: ""
spec:
  volumes:
  - name: fred
    volumesource:
      hostpath:
        path: /blee
        type: Directory
      emptydir: null
      gcepersistentdisk: null
      awselasticblockstore: null
      gitrepo: null
      secret: null
      nfs: null
      iscsi: null
      glusterfs: null
      persistentvolumeclaim: null
      rbd: null
      flexvolume: null
      cinder: null
      cephfs: null
      flocker: null
      downwardapi: null
      fc: null
      azurefile: null
      configmap: null
      vspherevolume: null
      quobyte: null
      azuredisk: null
      photonpersistentdisk: null
      projected: null
      portworxvolume: null
      scaleio: null
      storageos: null
  initcontainers: []
  containers:
  - name: fred
    image: blee
    command: []
    args: []
    workingdir: ""
    ports: []
    envfrom: []
    env:
    - name: fred
      value: "1"
      valuefrom:
        fieldref: null
        resourcefieldref: null
        configmapkeyref:
          localobjectreference:
            name: ""
          key: blee
          optional: null
        secretkeyref: null
    resources:
      limits: {}
      requests: {}
    volumemounts: []
    volumedevices: []
    livenessprobe: null
    readinessprobe: null
    lifecycle: null
    terminationmessagepath: ""
    terminationmessagepolicy: ""
    imagepullpolicy: ""
    securitycontext: null
    stdin: false
    stdinonce: false
    tty: false
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
  priorityclassname: bozo
  priority: 1
  dnsconfig: null
  readinessgates: []
  runtimeclassname: null
  enableservicelinks: null
status:
  phase: Running
  conditions: []
  message: ""
  reason: ""
  nominatednodename: ""
  hostip: ""
  podip: ""
  starttime: null
  initcontainerstatuses: []
  containerstatuses:
  - name: fred
    state:
      waiting: null
      running:
        startedat: "0001-01-01T00:00:00Z"
      terminated: null
    lastterminationstate:
      waiting: null
      running: null
      terminated: null
    ready: false
    restartcount: 0
    image: ""
    imageid: ""
    containerid: ""
  qosclass: ""
`
}
