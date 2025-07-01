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
	"github.com/stretchr/testify/require"
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
	c.Init(makePodFactory(), client.CoGVR)

	ctx := context.WithValue(context.Background(), internal.KeyPath, "fred/p1")
	oo, err := c.List(ctx, "")
	require.NoError(t, err)
	assert.Len(t, oo, 1)
}

// ----------------------------------------------------------------------------
// Helpers...

type conn struct{}

func makeConn() *conn {
	return &conn{}
}

func (*conn) Config() *client.Config                                   { return nil }
func (*conn) Dial() (kubernetes.Interface, error)                      { return nil, nil }
func (*conn) DialLogs() (kubernetes.Interface, error)                  { return nil, nil }
func (*conn) ConnectionOK() bool                                       { return true }
func (*conn) SwitchContext(string) error                               { return nil }
func (*conn) CachedDiscovery() (*disk.CachedDiscoveryClient, error)    { return nil, nil }
func (*conn) RestConfig() (*restclient.Config, error)                  { return nil, nil }
func (*conn) MXDial() (*versioned.Clientset, error)                    { return nil, nil }
func (*conn) DynDial() (dynamic.Interface, error)                      { return nil, nil }
func (*conn) HasMetrics() bool                                         { return false }
func (*conn) CheckConnectivity() bool                                  { return false }
func (*conn) IsNamespaced(string) bool                                 { return false }
func (*conn) SupportsResource(string) bool                             { return false }
func (*conn) ValidNamespaces() ([]v1.Namespace, error)                 { return nil, nil }
func (*conn) SupportsRes(string, []string) (a string, b bool, e error) { return "", false, nil }
func (*conn) ServerVersion() (*version.Info, error)                    { return nil, nil }
func (*conn) CurrentNamespaceName() (string, error)                    { return "", nil }
func (*conn) CanI(string, *client.GVR, string, []string) (bool, error) { return true, nil }
func (*conn) ActiveContext() string                                    { return "" }
func (*conn) ActiveNamespace() string                                  { return "" }
func (*conn) IsValidNamespace(string) bool                             { return true }
func (*conn) ValidNamespaceNames() (client.NamespaceNames, error)      { return nil, nil }
func (*conn) IsActiveNamespace(string) bool                            { return false }

type podFactory struct{}

var _ dao.Factory = &testFactory{}

func (podFactory) Client() client.Connection {
	return makeConn()
}

func (podFactory) Get(*client.GVR, string, bool, labels.Selector) (runtime.Object, error) {
	var m map[string]any
	if err := yaml.Unmarshal([]byte(poYaml()), &m); err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: m}, nil
}

func (podFactory) List(*client.GVR, string, bool, labels.Selector) ([]runtime.Object, error) {
	return nil, nil
}
func (podFactory) ForResource(string, *client.GVR) (informers.GenericInformer, error) {
	return nil, nil
}
func (podFactory) CanForResource(string, *client.GVR, []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (podFactory) WaitForCacheSync()            {}
func (podFactory) Forwarders() watch.Forwarders { return nil }
func (podFactory) DeleteForwarder(string)       {}

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
