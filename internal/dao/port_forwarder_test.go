// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/port"
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
)

// ----------------------------------------------------------------------------
// PortForwarder Authorization Test Suite
// ----------------------------------------------------------------------------

func TestPortForwarder_Start(t *testing.T) {
	cases := []struct {
		name            string
		phase           v1.PodPhase
		auth            authConfig
		wantAuthErr     bool   // true if error must contain auth denial
		wantErrContains string // expected error substring (empty = just require error)
	}{
		{
			name:            "pod not running",
			phase:           v1.PodPending,
			auth:            authConfig{canGetPod: true},
			wantAuthErr:     false,
			wantErrContains: "pod is not running",
		},
		{
			name:            "unauthorized to get pod",
			phase:           v1.PodRunning,
			auth:            authConfig{canGetPod: false},
			wantAuthErr:     false,
			wantErrContains: "not authorized to get pods",
		},
		{
			name:        "create only — legacy SPDY allowed",
			phase:       v1.PodRunning,
			auth:        authConfig{canGetPod: true, canCreatePortForward: true},
			wantAuthErr: false,
		},
		{
			name:        "get only — WebSocket path allowed",
			phase:       v1.PodRunning,
			auth:        authConfig{canGetPod: true, canGetPortForward: true},
			wantAuthErr: false,
		},
		{
			name:        "both permissions",
			phase:       v1.PodRunning,
			auth:        authConfig{canGetPod: true, canCreatePortForward: true, canGetPortForward: true},
			wantAuthErr: false,
		},
		{
			name:            "neither permission",
			phase:           v1.PodRunning,
			auth:            authConfig{canGetPod: true},
			wantAuthErr:     true,
			wantErrContains: "not authorized to access portforward",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pf := NewPortForwarder(makePortForwardFactoryWithAuth(tc.phase, tc.auth))
			_, err := pf.Start("test-ns/pod-1", testTunnel())
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrContains)
			if tc.wantAuthErr {
				assert.Contains(t, err.Error(), "not authorized to access portforward")
			} else {
				assert.NotContains(t, err.Error(), "not authorized to access portforward")
			}
		})
	}
}

// Removed separate per-scenario test functions in favor of table-driven test above.

// ----------------------------------------------------------------------------
// Test Helpers
// ----------------------------------------------------------------------------

type authConfig struct {
	canGetPod            bool
	canCreatePortForward bool
	canGetPortForward    bool
}

func testTunnel() port.PortTunnel {
	return port.NewPortTunnel("localhost", "main", "8080", "80")
}

func makePortForwardFactoryWithAuth(phase v1.PodPhase, cfg authConfig) Factory {
	return &pfFactory{authCfg: cfg, podPhase: phase}
}

func makePortForwardFactory(phase v1.PodPhase) Factory {
	return &pfFactory{
		authCfg: authConfig{
			canGetPod:            true,
			canCreatePortForward: true,
			canGetPortForward:    true,
		},
		podPhase: phase,
	}
}

type pfFactory struct {
	authCfg  authConfig
	podPhase v1.PodPhase
}

func (f *pfFactory) Client() client.Connection { return &pfConn{cfg: f.authCfg, phase: f.podPhase} }

func (f *pfFactory) Get(_ *client.GVR, path string, _ bool, _ labels.Selector) (runtime.Object, error) {
	ns, n := client.Namespaced(path)
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      n,
				"namespace": ns,
			},
			"status": map[string]interface{}{
				"phase": string(f.podPhase),
			},
		},
	}, nil
}

func (f *pfFactory) List(_ *client.GVR, _ string, _ bool, _ labels.Selector) ([]runtime.Object, error) {
	return nil, nil
}
func (f *pfFactory) ForResource(_ string, _ *client.GVR) (informers.GenericInformer, error) {
	return nil, nil
}
func (f *pfFactory) CanForResource(_ string, _ *client.GVR, _ []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (_ *pfFactory) WaitForCacheSync()            {}
func (_ *pfFactory) Forwarders() watch.Forwarders { return nil }
func (_ *pfFactory) DeleteForwarder(string)       {}

// ----------------------------------------------------------------------------
// Mock Connection
// ----------------------------------------------------------------------------

type pfConn struct {
	cfg     authConfig
	phase   v1.PodPhase
	dialErr error
}

func (*pfConn) Config() *client.Config                                { return nil }
func (c *pfConn) Dial() (kubernetes.Interface, error)                 { return nil, c.dialErr }
func (*pfConn) DialLogs() (kubernetes.Interface, error)               { return nil, nil }
func (*pfConn) ConnectionOK() bool                                    { return true }
func (*pfConn) SwitchContext(string) error                            { return nil }
func (*pfConn) CachedDiscovery() (*disk.CachedDiscoveryClient, error) { return nil, nil }
func (*pfConn) RestConfig() (*restclient.Config, error) {
	return nil, fmt.Errorf("mock: no real cluster")
}
func (*pfConn) MXDial() (*versioned.Clientset, error)                    { return nil, nil }
func (*pfConn) DynDial() (dynamic.Interface, error)                      { return nil, nil }
func (*pfConn) HasMetrics() bool                                         { return false }
func (*pfConn) CheckConnectivity() bool                                  { return false }
func (*pfConn) IsNamespaced(string) bool                                 { return false }
func (*pfConn) SupportsResource(string) bool                             { return false }
func (*pfConn) ValidNamespaces() ([]v1.Namespace, error)                 { return nil, nil }
func (*pfConn) SupportsRes(string, []string) (a string, b bool, e error) { return "", false, nil }
func (*pfConn) ServerVersion() (*version.Info, error)                    { return nil, nil }
func (*pfConn) CurrentNamespaceName() (string, error)                    { return "", nil }
func (*pfConn) ActiveContext() string                                    { return "" }
func (*pfConn) ActiveNamespace() string                                  { return "" }
func (*pfConn) IsValidNamespace(string) bool                             { return true }
func (*pfConn) ValidNamespaceNames() (client.NamespaceNames, error)      { return nil, nil }
func (*pfConn) IsActiveNamespace(string) bool                            { return false }

func (c *pfConn) CanI(ns string, gvr *client.GVR, name string, verbs []string) (bool, error) {
	gvrStr := gvr.String()

	if gvrStr == client.PodGVR.String() {
		for _, v := range verbs {
			if v == client.GetVerb && c.cfg.canGetPod {
				return true, nil
			}
		}
		return false, nil
	}

	if gvrStr == client.PodGVR.WithSubResource("portforward").String() {
		for _, v := range verbs {
			if v == client.CreateVerb && c.cfg.canCreatePortForward {
				return true, nil
			}
			if v == client.GetVerb && c.cfg.canGetPortForward {
				return true, nil
			}
		}
		return false, nil
	}

	return false, nil
}
