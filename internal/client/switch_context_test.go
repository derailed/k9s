// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// newFakeK8sServer creates a minimal fake Kubernetes API server
// that responds to version and discovery endpoints.
// Returns the server and a counter for /version calls.
func newFakeK8sServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	versionCalls := &atomic.Int32{}
	mux := http.NewServeMux()

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		versionCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(version.Info{
			Major:      "1",
			Minor:      "28",
			GitVersion: "v1.28.0",
		})
	})

	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"kind":"APIVersions","versions":["v1"]}`))
	})

	mux.HandleFunc("/apis", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"kind":"APIGroupList","apiVersion":"v1","groups":[]}`))
	})

	return httptest.NewServer(mux), versionCalls
}

// writeSwitchTestKubeconfig creates a temporary kubeconfig with two contexts
// pointing to the given server URLs.
func writeSwitchTestKubeconfig(t *testing.T, server1URL, server2URL string) string {
	t.Helper()

	dir := t.TempDir()
	kubeconfig := filepath.Join(dir, "config")
	content := `apiVersion: v1
kind: Config
current-context: context1
contexts:
- context:
    cluster: cluster1
    user: user1
    namespace: default
  name: context1
- context:
    cluster: cluster2
    user: user1
    namespace: kube-system
  name: context2
clusters:
- cluster:
    server: ` + server1URL + `
  name: cluster1
- cluster:
    server: ` + server2URL + `
  name: cluster2
users:
- name: user1
  user:
    token: test-token
`
	require.NoError(t, os.WriteFile(kubeconfig, []byte(content), 0644))
	return kubeconfig
}

// newSwitchTestClient creates an APIClient configured for context-switch
// testing with the given kubeconfig and initial context.
func newSwitchTestClient(t *testing.T, kubeconfig, ctx string) *APIClient {
	t.Helper()

	flags := genericclioptions.NewConfigFlags(false)
	flags.KubeConfig = &kubeconfig
	flags.Context = &ctx

	return &APIClient{
		config: NewConfig(flags),
		cache:  cache.NewLRUExpireCache(cacheSize),
		connOK: true,
		log:    slog.Default(),
	}
}

func TestSwitchContextSuccess(t *testing.T) {
	srv, _ := newFakeK8sServer(t)
	defer srv.Close()
	t.Setenv("HOME", t.TempDir())

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	a := newSwitchTestClient(t, kubeconfig, "context1")

	err := a.SwitchContext("context2")
	require.NoError(t, err)

	ctx, err := a.config.CurrentContextName()
	require.NoError(t, err)
	assert.Equal(t, "context2", ctx)
	assert.True(t, a.getConnOK())
}

func TestSwitchContextReusesConnectivityClient(t *testing.T) {
	srv, _ := newFakeK8sServer(t)
	defer srv.Close()
	t.Setenv("HOME", t.TempDir())

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	a := newSwitchTestClient(t, kubeconfig, "context1")

	err := a.SwitchContext("context2")
	require.NoError(t, err)

	assert.NotNil(t, a.getClient(),
		"SwitchContext should store the connectivity-check client for reuse by Dial()")
}

func TestSwitchContextPreWarmsDynDial(t *testing.T) {
	srv, _ := newFakeK8sServer(t)
	defer srv.Close()
	t.Setenv("HOME", t.TempDir())

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	a := newSwitchTestClient(t, kubeconfig, "context1")

	err := a.SwitchContext("context2")
	require.NoError(t, err)

	assert.NotNil(t, a.getDClient(),
		"SwitchContext should pre-warm the dynamic client so gotoResource reuses it")
}

func TestSwitchContextDialAfterSwitch(t *testing.T) {
	srv, _ := newFakeK8sServer(t)
	defer srv.Close()
	t.Setenv("HOME", t.TempDir())

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	a := newSwitchTestClient(t, kubeconfig, "context1")

	err := a.SwitchContext("context2")
	require.NoError(t, err)

	storedClient := a.getClient()
	require.NotNil(t, storedClient, "SwitchContext should store client")

	dialedClient, err := a.Dial()
	require.NoError(t, err)
	assert.Same(t, storedClient, dialedClient,
		"Dial() should return the stored connectivity client, not create a new one")
}

func TestSwitchContextMinimalVersionCalls(t *testing.T) {
	srv, versionCalls := newFakeK8sServer(t)
	defer srv.Close()
	t.Setenv("HOME", t.TempDir())

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	a := newSwitchTestClient(t, kubeconfig, "context1")

	err := a.SwitchContext("context2")
	require.NoError(t, err)

	assert.Equal(t, int32(1), versionCalls.Load(),
		"SwitchContext should call ServerVersion exactly once")
}

func TestSwitchContextInvalidContext(t *testing.T) {
	srv, _ := newFakeK8sServer(t)
	defer srv.Close()

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	a := newSwitchTestClient(t, kubeconfig, "context1")

	err := a.SwitchContext("nonexistent")
	assert.Error(t, err)
}


func TestInitConnectionMetricsUnsupported(t *testing.T) {
	srv, _ := newFakeK8sServer(t)
	defer srv.Close()
	t.Setenv("HOME", t.TempDir())

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	flags := genericclioptions.NewConfigFlags(false)
	flags.KubeConfig = &kubeconfig
	ctx := "context1"
	flags.Context = &ctx
	cfg := NewConfig(flags)

	a, err := InitConnection(cfg, slog.Default())
	require.NoError(t, err)
	assert.True(t, a.ConnectionOK(),
		"InitConnection should succeed when metrics-server is absent")
}

func TestInitConnectionStoresDialClient(t *testing.T) {
	srv, _ := newFakeK8sServer(t)
	defer srv.Close()
	t.Setenv("HOME", t.TempDir())

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	flags := genericclioptions.NewConfigFlags(false)
	flags.KubeConfig = &kubeconfig
	ctx := "context1"
	flags.Context = &ctx
	cfg := NewConfig(flags)

	a, err := InitConnection(cfg, slog.Default())
	require.NoError(t, err)

	assert.NotNil(t, a.getClient(),
		"InitConnection should store a Dial client for reuse")
}