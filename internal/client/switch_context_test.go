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

const (
	testContext1 = "context1"
	testContext2 = "context2"
)

func newFakeK8sServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	versionCalls := &atomic.Int32{}
	mux := http.NewServeMux()

	mux.HandleFunc("/version", func(w http.ResponseWriter, _ *http.Request) {
		versionCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(version.Info{
			Major:      "1",
			Minor:      "28",
			GitVersion: "v1.28.0",
		})
	})

	mux.HandleFunc("/api", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"kind":"APIVersions","versions":["v1"]}`))
	})

	mux.HandleFunc("/apis", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"kind":"APIGroupList","apiVersion":"v1","groups":[]}`))
	})

	return httptest.NewServer(mux), versionCalls
}

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
	require.NoError(t, os.WriteFile(kubeconfig, []byte(content), 0600))

	return kubeconfig
}

func setupSwitchTest(t *testing.T) (*httptest.Server, *atomic.Int32, *APIClient) {
	t.Helper()

	srv, versionCalls := newFakeK8sServer(t)
	t.Cleanup(srv.Close)
	t.Setenv("HOME", t.TempDir())

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	flags := genericclioptions.NewConfigFlags(false)
	flags.KubeConfig = &kubeconfig
	ctx := testContext1
	flags.Context = &ctx

	a := &APIClient{
		config: NewConfig(flags),
		cache:  cache.NewLRUExpireCache(cacheSize),
		connOK: true,
		log:    slog.Default(),
	}

	return srv, versionCalls, a
}

func TestSwitchContextSuccess(t *testing.T) {
	_, _, a := setupSwitchTest(t)

	err := a.SwitchContext(testContext2)
	require.NoError(t, err)

	ctx, err := a.config.CurrentContextName()
	require.NoError(t, err)
	assert.Equal(t, testContext2, ctx)
	assert.True(t, a.getConnOK())
}

func TestSwitchContextReusesConnectivityClient(t *testing.T) {
	_, _, a := setupSwitchTest(t)

	err := a.SwitchContext(testContext2)
	require.NoError(t, err)

	assert.NotNil(t, a.getClient(),
		"SwitchContext should store the connectivity-check client for reuse by Dial()")
}

func TestSwitchContextPreWarmsDynDial(t *testing.T) {
	_, _, a := setupSwitchTest(t)

	err := a.SwitchContext(testContext2)
	require.NoError(t, err)

	assert.NotNil(t, a.getDClient(),
		"SwitchContext should pre-warm the dynamic client so gotoResource reuses it")
}

func TestSwitchContextDialAfterSwitch(t *testing.T) {
	_, _, a := setupSwitchTest(t)

	err := a.SwitchContext(testContext2)
	require.NoError(t, err)

	storedClient := a.getClient()
	require.NotNil(t, storedClient, "SwitchContext should store client")

	dialedClient, err := a.Dial()
	require.NoError(t, err)
	assert.Same(t, storedClient, dialedClient,
		"Dial() should return the stored connectivity client, not create a new one")
}

func TestSwitchContextMinimalVersionCalls(t *testing.T) {
	_, versionCalls, a := setupSwitchTest(t)

	err := a.SwitchContext(testContext2)
	require.NoError(t, err)

	assert.Equal(t, int32(1), versionCalls.Load(),
		"SwitchContext should call ServerVersion exactly once")
}

func TestSwitchContextInvalidContext(t *testing.T) {
	_, _, a := setupSwitchTest(t)

	err := a.SwitchContext("nonexistent")
	assert.Error(t, err)
}

func TestInitConnectionMetricsUnsupported(t *testing.T) {
	srv, _, _ := setupSwitchTest(t)

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	flags := genericclioptions.NewConfigFlags(false)
	flags.KubeConfig = &kubeconfig
	ctx := testContext1
	flags.Context = &ctx

	a, err := InitConnection(NewConfig(flags), slog.Default())
	require.NoError(t, err)
	assert.True(t, a.ConnectionOK(),
		"InitConnection should succeed when metrics-server is absent")
}

func TestInitConnectionStoresDialClient(t *testing.T) {
	srv, _, _ := setupSwitchTest(t)

	kubeconfig := writeSwitchTestKubeconfig(t, srv.URL, srv.URL)
	flags := genericclioptions.NewConfigFlags(false)
	flags.KubeConfig = &kubeconfig
	ctx := testContext1
	flags.Context = &ctx

	a, err := InitConnection(NewConfig(flags), slog.Default())
	require.NoError(t, err)

	assert.NotNil(t, a.getClient(),
		"InitConnection should store a Dial client for reuse")
}
