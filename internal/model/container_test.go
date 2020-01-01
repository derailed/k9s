package model_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

func TestContainerList(t *testing.T) {
	c := model.Container{}
	c.Init(render.ClusterScope, "containers", makePodFactory())

	ctx := context.WithValue(context.Background(), internal.KeyPath, "fred/p1")
	oo, err := c.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(oo))
}

func TestContainerHydrate(t *testing.T) {
	c := model.Container{}
	c.Init(render.ClusterScope, "containers", makePodFactory())

	ctx := context.WithValue(context.Background(), internal.KeyPath, "fred/p1")
	oo, err := c.List(ctx)
	assert.Nil(t, err)

	rr := make(render.Rows, len(oo))
	assert.Nil(t, c.Hydrate(oo, rr, render.Container{}))
	assert.Equal(t, 1, len(rr))
	assert.Equal(t, "fred", rr[0].ID)
	assert.Equal(t, render.Fields{"fred", "blee", "false", "Running", "false", "0", "off:off", "n/a", "n/a", "n/a", "n/a", ""}, rr[0].Fields[0:len(rr[0].Fields)-1])
}

// ----------------------------------------------------------------------------
// Helpers...

type podFactory struct{}

var _ dao.Factory = testFactory{}

func (f podFactory) Client() client.Connection {
	return nil
}
func (f podFactory) Get(gvr, path string, wait bool, sel labels.Selector) (runtime.Object, error) {
	var m map[string]interface{}
	if err := yaml.Unmarshal([]byte(poYaml()), &m); err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: m}, nil
}
func (f podFactory) List(gvr, ns string, wait bool, sel labels.Selector) ([]runtime.Object, error) {
	return nil, nil
}
func (f podFactory) ForResource(ns, gvr string) informers.GenericInformer { return nil }
func (f podFactory) CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (f podFactory) WaitForCacheSync()            {}
func (f podFactory) Forwarders() watch.Forwarders { return nil }
func (f podFactory) DeleteForwarder(string)       {}

func makePodFactory() dao.Factory {
	return podFactory{}
}

func poYaml() string {
	return `apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  labels:
    blee: duh
  name: fred
  namespace: blee
spec:
  containers:
  - env:
    - name: fred
      value: "1"
      valueFrom:
        configMapKeyRef:
          key: blee
    image: blee
    name: fred
    resources: {}
  priority: 1
  priorityClassName: bozo
  volumes:
  - hostPath:
      path: /blee
      type: Directory
    name: fred
status:
  containerStatuses:
  - image: ""
    imageID: ""
    lastState: {}
    name: fred
    ready: false
    restartCount: 0
    state:
      running:
        startedAt: null
  phase: Running
`
}
