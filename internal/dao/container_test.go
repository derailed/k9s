// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/yaml"
)

func TestContainerList(t *testing.T) {
	c := dao.Container{}
	c.Init(makePodFactory(), client.NewGVR("containers"))

	ctx := context.WithValue(context.Background(), internal.KeyPath, "fred/p1")
	oo, err := c.List(ctx, "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(oo))
}

// ----------------------------------------------------------------------------
// Helpers...

type conn struct{}

func makeConn() *conn {
	return &conn{}
}

func (c *conn) Config() *client.Config                                { return nil }
func (c *conn) Dial() (kubernetes.Interface, error)                   { return nil, nil }
func (c *conn) DialLogs() (kubernetes.Interface, error)               { return nil, nil }
func (c *conn) ConnectionOK() bool                                    { return true }
func (c *conn) SwitchContext(ctx string) error                        { return nil }
func (c *conn) CachedDiscovery() (*disk.CachedDiscoveryClient, error) { return nil, nil }
func (c *conn) RestConfig() (*restclient.Config, error)               { return nil, nil }
func (c *conn) MXDial() (*versioned.Clientset, error)                 { return nil, nil }
func (c *conn) DynDial() (dynamic.Interface, error)                   { return nil, nil }
func (c *conn) HasMetrics() bool                                      { return false }
func (c *conn) CheckConnectivity() bool                               { return false }
func (c *conn) IsNamespaced(n string) bool                            { return false }
func (c *conn) SupportsResource(group string) bool                    { return false }
func (c *conn) ValidNamespaces() ([]v1.Namespace, error)              { return nil, nil }
func (c *conn) SupportsRes(grp string, versions []string) (string, bool, error) {
	return "", false, nil
}
func (c *conn) ServerVersion() (*version.Info, error)                { return nil, nil }
func (c *conn) CurrentNamespaceName() (string, error)                { return "", nil }
func (c *conn) CanI(ns, gvr, n string, verbs []string) (bool, error) { return true, nil }
func (c *conn) ActiveContext() string                                { return "" }
func (c *conn) ActiveNamespace() string                              { return "" }
func (c *conn) IsValidNamespace(string) bool                         { return true }
func (c *conn) ValidNamespaceNames() (client.NamespaceNames, error)  { return nil, nil }
func (c *conn) IsActiveNamespace(string) bool                        { return false }

type podFactory struct{}

var _ dao.Factory = &testFactory{}

func (f podFactory) Client() client.Connection {
	return makeConn()
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
func (f podFactory) ForResource(ns, gvr string) (informers.GenericInformer, error) { return nil, nil }
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
