// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package watch

import (
	"log/slog"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	version "k8s.io/apimachinery/pkg/version"
	disk "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	dfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

// TestCanForResourceNsGVRUsesBlankNamespace verifies that CanForResource for the
// namespaces GVR always creates a cluster-scoped informer factory (keyed by
// BlankNamespace) regardless of the active namespace. Before the fix, passing
// ns="default" caused ensureFactory to create a namespace-scoped factory, which
// in turn built the invalid API path /api/v1/namespaces/default/namespaces.
func TestCanForResourceNsGVRUsesBlankNamespace(t *testing.T) {
	uu := map[string]struct {
		ns string
	}{
		"active-namespace": {ns: "default"},
		"other-namespace":  {ns: "kube-system"},
	}

	for name, u := range uu {
		t.Run(name, func(t *testing.T) {
			conn := newMockConn()
			f := NewFactory(conn)

			_, err := f.CanForResource(u.ns, client.NsGVR, client.ListAccess)
			require.NoError(t, err)

			// The factory for the namespaces GVR must be keyed by BlankNamespace,
			// not by the active namespace, so the API path stays cluster-scoped.
			_, hasBlank := f.factories[client.BlankNamespace]
			assert.True(t, hasBlank, "expected cluster-scoped factory (BlankNamespace key)")

			_, hasNs := f.factories[u.ns]
			assert.False(t, hasNs, "unexpected namespace-scoped factory for %q — would produce invalid API path /api/v1/namespaces/%s/namespaces", u.ns, u.ns)
		})
	}
}

// mockConn satisfies client.Connection for tests; DynDial returns a no-op fake.
type mockConn struct{}

func newMockConn() mockConn { return mockConn{} }

func (mockConn) CanI(string, *client.GVR, string, []string) (bool, error) { return true, nil }
func (mockConn) Config() *client.Config                                    { return nil }
func (mockConn) ConnectionOK() bool                                        { return true }
func (mockConn) Dial() (kubernetes.Interface, error)                       { return nil, nil }
func (mockConn) DialLogs() (kubernetes.Interface, error)                   { return nil, nil }
func (mockConn) SwitchContext(string) error                                { return nil }
func (mockConn) CachedDiscovery() (*disk.CachedDiscoveryClient, error)    { return nil, nil }
func (mockConn) RestConfig() (*restclient.Config, error)                   { return nil, nil }
func (mockConn) MXDial() (*versioned.Clientset, error)                    { return nil, nil }
func (mockConn) DynDial() (dynamic.Interface, error) {
	return dfake.NewSimpleDynamicClient(scheme.Scheme), nil
}
func (mockConn) HasMetrics() bool                                 { return false }
func (mockConn) ValidNamespaceNames() (client.NamespaceNames, error) { return nil, nil }
func (mockConn) IsValidNamespace(string) bool                     { return true }
func (mockConn) ServerVersion() (*version.Info, error)            { return nil, nil }
func (mockConn) CheckConnectivity() bool                          { return false }
func (mockConn) ActiveContext() string                            { return "" }
func (mockConn) ActiveNamespace() string                          { return "" }
func (mockConn) IsActiveNamespace(string) bool                    { return false }
