// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	helmtime "helm.sh/helm/v3/pkg/time"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"
)

const (
	helmWait = 10 * time.Second
	helmTick = 20 * time.Millisecond
)

type helmConn struct {
	client.Connection

	cfg *client.Config
}

func (c *helmConn) Config() *client.Config { return c.cfg }

func (c *helmConn) RestConfig() (*restclient.Config, error) { return c.cfg.RESTConfig() }

type helmFactory struct {
	Factory

	conn client.Connection
}

func (f *helmFactory) Client() client.Connection { return f.conn }

// helmSecret describes a release revision stored by the fake API server.
type helmRevSecret struct {
	ns, name string
	version  int
	status   string
	rv       string
}

func (s *helmRevSecret) secretName() string {
	return fmt.Sprintf("sh.helm.release.v1.%s.v%d", s.name, s.version)
}

func (s *helmRevSecret) uid() string {
	return fmt.Sprintf("uid-%s-%s-%d", s.ns, s.name, s.version)
}

func (s *helmRevSecret) release() *release.Release {
	return &release.Release{
		Name:      s.name,
		Namespace: s.ns,
		Version:   s.version,
		Info: &release.Info{
			Status:       release.Status(s.status),
			LastDeployed: helmtime.Time{Time: time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)},
		},
		Chart: &chart.Chart{Metadata: &chart.Metadata{
			Name:       "chart-" + s.name,
			Version:    "1." + strconv.Itoa(s.version) + ".0",
			AppVersion: "app-" + s.name,
		}},
	}
}

// encode mirrors helm's secrets driver release encoding.
func (s *helmRevSecret) encode(t *testing.T) string {
	t.Helper()

	b, err := json.Marshal(s.release())
	require.NoError(t, err)
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	require.NoError(t, err)
	_, err = w.Write(b)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func (s *helmRevSecret) objectMeta() map[string]any {
	return map[string]any{
		"name":              s.secretName(),
		"namespace":         s.ns,
		"uid":               s.uid(),
		"resourceVersion":   s.rv,
		"creationTimestamp": "2026-03-04T05:06:07Z",
		"labels": map[string]string{
			"owner":   "helm",
			"name":    s.name,
			"status":  s.status,
			"version": strconv.Itoa(s.version),
		},
	}
}

// helmAPI is a minimal fake API server serving helm release secrets.
type helmAPI struct {
	t   *testing.T
	srv *httptest.Server

	mx          sync.Mutex
	secrets     []*helmRevSecret
	metaStatus  int
	failGets    map[string]bool
	metaReqs    []*url.URL
	metaAccepts []string
	fullReqs    []string
	cmReqs      int
	gets        map[string]int
}

const helmPage = 3

func newHelmAPI(t *testing.T) *helmAPI {
	t.Helper()

	a := &helmAPI{
		t:        t,
		failGets: make(map[string]bool),
		gets:     make(map[string]int),
		secrets: []*helmRevSecret{
			{ns: "ns1", name: "app", version: 1, status: "superseded", rv: "10"},
			{ns: "ns1", name: "app", version: 3, status: "deployed", rv: "30"},
			{ns: "ns1", name: "other", version: 1, status: "failed", rv: "50"},
			{ns: "ns1", name: "app", version: 2, status: "superseded", rv: "20"},
			{ns: "ns2", name: "app", version: 1, status: "deployed", rv: "40"},
		},
	}
	a.srv = httptest.NewServer(http.HandlerFunc(a.serve))
	t.Cleanup(a.srv.Close)

	return a
}

func (a *helmAPI) writeJSON(w http.ResponseWriter, code int, o any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(o); err != nil {
		a.t.Errorf("fake api encode failed: %v", err)
	}
}

func (a *helmAPI) serve(w http.ResponseWriter, r *http.Request) {
	a.mx.Lock()
	defer a.mx.Unlock()

	p := r.URL.Path
	switch {
	case p == "/version":
		a.writeJSON(w, http.StatusOK, map[string]string{"major": "1", "minor": "30", "gitVersion": "v1.30.0"})
	case p == "/api" || p == "/apis":
		a.writeJSON(w, http.StatusOK, map[string]any{"kind": "APIVersions", "versions": []string{"v1"}})
	case p == "/api/v1/configmaps" || strings.HasSuffix(p, "/configmaps"):
		a.cmReqs++
		a.writeJSON(w, http.StatusOK, map[string]any{
			"apiVersion": "v1", "kind": "ConfigMapList", "items": []any{},
		})
	case p == "/api/v1/secrets" || (strings.HasPrefix(p, "/api/v1/namespaces/") && strings.HasSuffix(p, "/secrets")):
		a.listSecrets(w, r)
	case strings.HasPrefix(p, "/api/v1/namespaces/") && strings.Contains(p, "/secrets/"):
		a.getSecret(w, r)
	default:
		a.t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		http.NotFound(w, r)
	}
}

func (a *helmAPI) scoped(p string) []*helmRevSecret {
	if p == "/api/v1/secrets" {
		return a.secrets
	}
	ns := strings.TrimSuffix(strings.TrimPrefix(p, "/api/v1/namespaces/"), "/secrets")
	ss := make([]*helmRevSecret, 0, len(a.secrets))
	for _, s := range a.secrets {
		if s.ns == ns {
			ss = append(ss, s)
		}
	}

	return ss
}

func (a *helmAPI) listSecrets(w http.ResponseWriter, r *http.Request) {
	accept := r.Header.Get("Accept")
	if !strings.Contains(accept, "PartialObjectMetadataList") {
		a.fullReqs = append(a.fullReqs, r.URL.Path)
		items := make([]map[string]any, 0, len(a.secrets))
		for _, s := range a.scoped(r.URL.Path) {
			items = append(items, map[string]any{
				"metadata": s.objectMeta(),
				"type":     "helm.sh/release.v1",
				"data":     map[string]string{"release": base64.StdEncoding.EncodeToString([]byte(s.encode(a.t)))},
			})
		}
		a.writeJSON(w, http.StatusOK, map[string]any{"apiVersion": "v1", "kind": "SecretList", "items": items})
		return
	}

	u := *r.URL
	a.metaReqs = append(a.metaReqs, &u)
	a.metaAccepts = append(a.metaAccepts, accept)
	if a.metaStatus != 0 {
		a.writeJSON(w, a.metaStatus, map[string]any{
			"apiVersion": "v1", "kind": "Status", "status": "Failure", "code": a.metaStatus,
			"reason": "InternalError", "message": "boom",
		})
		return
	}

	ss := a.scoped(r.URL.Path)
	start := 0
	if c := r.URL.Query().Get("continue"); c != "" {
		start, _ = strconv.Atoi(c)
	}
	end, next := min(start+helmPage, len(ss)), ""
	if end < len(ss) {
		next = strconv.Itoa(end)
	}
	items := make([]map[string]any, 0, helmPage)
	for _, s := range ss[start:end] {
		items = append(items, map[string]any{
			"apiVersion": "v1", "kind": "Secret", "metadata": s.objectMeta(),
		})
	}
	a.writeJSON(w, http.StatusOK, map[string]any{
		"apiVersion": "meta.k8s.io/v1",
		"kind":       "PartialObjectMetadataList",
		"metadata":   map[string]string{"continue": next},
		"items":      items,
	})
}

func (a *helmAPI) getSecret(w http.ResponseWriter, r *http.Request) {
	tokens := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/namespaces/"), "/secrets/")
	ns, name := tokens[0], tokens[1]
	key := ns + "/" + name
	a.gets[key]++
	if a.failGets[key] {
		a.writeJSON(w, http.StatusInternalServerError, map[string]any{
			"apiVersion": "v1", "kind": "Status", "status": "Failure", "code": 500,
		})
		return
	}
	for _, s := range a.secrets {
		if s.ns == ns && s.secretName() == name {
			a.writeJSON(w, http.StatusOK, map[string]any{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata":   s.objectMeta(),
				"type":       "helm.sh/release.v1",
				"data":       map[string]string{"release": base64.StdEncoding.EncodeToString([]byte(s.encode(a.t)))},
			})
			return
		}
	}
	http.NotFound(w, r)
}

func (a *helmAPI) getCount(key string) int {
	a.mx.Lock()
	defer a.mx.Unlock()

	return a.gets[key]
}

func (a *helmAPI) totalGets() int {
	a.mx.Lock()
	defer a.mx.Unlock()

	var n int
	for _, c := range a.gets {
		n += c
	}

	return n
}

func (a *helmAPI) metaCount() int {
	a.mx.Lock()
	defer a.mx.Unlock()

	return len(a.metaReqs)
}

func (a *helmAPI) bump(ns, name string, version int, rv string) {
	a.mx.Lock()
	defer a.mx.Unlock()

	for _, s := range a.secrets {
		if s.ns == ns && s.name == name && s.version == version {
			s.rv = rv
		}
	}
}

// newHelmChart returns a HelmChart wired to the fake API server and an empty cache.
func newHelmChart(t *testing.T, a *helmAPI) *HelmChart {
	t.Helper()

	cfgFile := filepath.Join(t.TempDir(), "kubeconfig")
	kubeCfg := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: t
  cluster:
    server: %s
users:
- name: t
  user:
    token: tok
contexts:
- name: t
  context:
    cluster: t
    user: t
current-context: t
`, a.srv.URL)
	require.NoError(t, os.WriteFile(cfgFile, []byte(kubeCfg), 0o600))
	t.Setenv("KUBECONFIG", cfgFile)
	t.Setenv("HELM_DRIVER", "")

	flags := genericclioptions.NewConfigFlags(false)
	flags.KubeConfig = &cfgFile

	helmCache = newHelmSummaries()
	cache := helmCache
	t.Cleanup(func() {
		// Let background fetches finish before the fake server goes away.
		assert.Eventually(t, func() bool {
			cache.mx.Lock()
			defer cache.mx.Unlock()

			return len(cache.inflight) == 0
		}, helmWait, helmTick)
	})

	h := new(HelmChart)
	h.Init(&helmFactory{conn: &helmConn{cfg: client.NewConfig(flags)}}, client.HmGVR)

	return h
}

func releases(t *testing.T, oo []runtime.Object) map[string]*release.Release {
	t.Helper()

	rr := make(map[string]*release.Release, len(oo))
	for _, o := range oo {
		r, ok := o.(helm.ReleaseRes)
		require.True(t, ok, "unexpected type %T", o)
		rr[client.FQN(r.Release.Namespace, r.Release.Name)] = r.Release
	}

	return rr
}

func allEnriched(t *testing.T, h *HelmChart, ns string, want int) bool {
	t.Helper()

	oo, err := h.List(context.Background(), ns)
	if err != nil || len(oo) != want {
		return false
	}
	for _, r := range releases(t, oo) {
		if r.Chart == nil {
			return false
		}
	}

	return true
}

func TestHelmChartListMetadata(t *testing.T) {
	a := newHelmAPI(t)
	h := newHelmChart(t, a)

	oo, err := h.List(context.Background(), client.NamespaceAll)
	require.NoError(t, err)

	// Only the latest revision of each release, without chart details yet.
	rr := releases(t, oo)
	require.Len(t, rr, 3)
	assert.Equal(t, 3, rr["ns1/app"].Version)
	assert.Equal(t, "deployed", rr["ns1/app"].Info.Status.String())
	assert.Equal(t, 1, rr["ns1/other"].Version)
	assert.Equal(t, "failed", rr["ns1/other"].Info.Status.String())
	assert.Equal(t, 1, rr["ns2/app"].Version)
	for fqn, r := range rr {
		assert.Nil(t, r.Chart, fqn)
	}

	// The list is a paginated, metadata only, owner=helm request.
	a.mx.Lock()
	require.Len(t, a.metaReqs, 2)
	for _, u := range a.metaReqs {
		assert.Equal(t, "/api/v1/secrets", u.Path)
		assert.Equal(t, "owner=helm", u.Query().Get("labelSelector"))
		assert.Equal(t, "500", u.Query().Get("limit"))
	}
	assert.Empty(t, a.metaReqs[0].Query().Get("continue"))
	assert.Equal(t, "3", a.metaReqs[1].Query().Get("continue"))
	assert.Empty(t, a.fullReqs)
	a.mx.Unlock()

	// Chart details are loaded in the background.
	require.Eventually(t, func() bool { return allEnriched(t, h, client.NamespaceAll, 3) }, helmWait, helmTick)
	oo, err = h.List(context.Background(), client.NamespaceAll)
	require.NoError(t, err)
	rr = releases(t, oo)
	assert.Equal(t, "chart-app", rr["ns1/app"].Chart.Metadata.Name)
	assert.Equal(t, "1.3.0", rr["ns1/app"].Chart.Metadata.Version)
	assert.Equal(t, "app-app", rr["ns1/app"].Chart.Metadata.AppVersion)
	assert.Equal(t, "chart-other", rr["ns1/other"].Chart.Metadata.Name)
	assert.Equal(t, time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC), rr["ns1/app"].Info.LastDeployed.Time)

	// Each latest revision payload was downloaded exactly once...
	assert.Equal(t, 1, a.getCount("ns1/sh.helm.release.v1.app.v3"))
	assert.Equal(t, 1, a.getCount("ns1/sh.helm.release.v1.other.v1"))
	assert.Equal(t, 1, a.getCount("ns2/sh.helm.release.v1.app.v1"))
	assert.Equal(t, 3, a.totalGets(), "old revisions must never be downloaded")

	// ...and further refreshes download no more payloads.
	for range 3 {
		_, err = h.List(context.Background(), client.NamespaceAll)
		require.NoError(t, err)
	}
	assert.Equal(t, 3, a.totalGets())
}

func TestHelmChartListNamespaced(t *testing.T) {
	a := newHelmAPI(t)
	h := newHelmChart(t, a)

	oo, err := h.List(context.Background(), "ns1")
	require.NoError(t, err)

	rr := releases(t, oo)
	require.Len(t, rr, 2)
	assert.Contains(t, rr, "ns1/app")
	assert.Contains(t, rr, "ns1/other")

	a.mx.Lock()
	defer a.mx.Unlock()
	require.NotEmpty(t, a.metaReqs)
	assert.Equal(t, "/api/v1/namespaces/ns1/secrets", a.metaReqs[0].Path)
}

func TestHelmChartListBlankNamespaceIsClusterWide(t *testing.T) {
	a := newHelmAPI(t)
	h := newHelmChart(t, a)

	oo, err := h.List(context.Background(), client.BlankNamespace)
	require.NoError(t, err)
	assert.Len(t, oo, 3)

	a.mx.Lock()
	defer a.mx.Unlock()
	require.NotEmpty(t, a.metaReqs)
	assert.Equal(t, "/api/v1/secrets", a.metaReqs[0].Path)
}

func TestHelmChartListRefetchesChangedRevision(t *testing.T) {
	a := newHelmAPI(t)
	h := newHelmChart(t, a)

	require.Eventually(t, func() bool { return allEnriched(t, h, client.NamespaceAll, 3) }, helmWait, helmTick)
	key := "ns1/sh.helm.release.v1.app.v3"
	require.Equal(t, 1, a.getCount(key))

	// A modified secret (e.g. pending-upgrade -> deployed) must be reloaded.
	a.bump("ns1", "app", 3, "31")
	oo, err := h.List(context.Background(), client.NamespaceAll)
	require.NoError(t, err)
	assert.Nil(t, releases(t, oo)["ns1/app"].Chart)

	require.Eventually(t, func() bool { return allEnriched(t, h, client.NamespaceAll, 3) }, helmWait, helmTick)
	assert.Equal(t, 2, a.getCount(key))
	assert.Equal(t, 1, a.getCount("ns1/sh.helm.release.v1.other.v1"))

	// The superseded cache entry is pruned.
	helmCache.mx.Lock()
	defer helmCache.mx.Unlock()
	assert.Len(t, helmCache.done, 3)
	_, stale := helmCache.done[helmCacheKey("uid-ns1-app-3", "30")]
	assert.False(t, stale)
}

func TestHelmChartListBacksOffOnFetchFailure(t *testing.T) {
	a := newHelmAPI(t)
	key := "ns1/sh.helm.release.v1.other.v1"
	a.failGets[key] = true
	h := newHelmChart(t, a)

	_, err := h.List(context.Background(), client.NamespaceAll)
	require.NoError(t, err)
	require.Eventually(t, func() bool { return a.getCount(key) >= 1 }, helmWait, helmTick)

	// Wait for the failure to be recorded, then make sure it is not retried.
	require.Eventually(t, func() bool {
		helmCache.mx.Lock()
		defer helmCache.mx.Unlock()

		return len(helmCache.failed) == 1 && len(helmCache.inflight) == 0
	}, helmWait, helmTick)
	before := a.getCount(key)
	for range 3 {
		oo, err := h.List(context.Background(), client.NamespaceAll)
		require.NoError(t, err)
		assert.Nil(t, releases(t, oo)["ns1/other"].Chart)
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, before, a.getCount(key))
}

func TestHelmChartListFallsBackOnMetadataError(t *testing.T) {
	a := newHelmAPI(t)
	a.metaStatus = http.StatusInternalServerError
	h := newHelmChart(t, a)

	oo, err := h.List(context.Background(), client.BlankNamespace)
	require.NoError(t, err)

	// Served by helm itself: full release details are available immediately.
	rr := releases(t, oo)
	require.Len(t, rr, 3)
	for fqn, r := range rr {
		assert.NotNil(t, r.Chart, fqn)
	}
	assert.Positive(t, a.metaCount())
	a.mx.Lock()
	defer a.mx.Unlock()
	assert.NotEmpty(t, a.fullReqs)
}

func TestHelmChartListSkipsMetadataForOtherDrivers(t *testing.T) {
	a := newHelmAPI(t)
	h := newHelmChart(t, a)
	t.Setenv("HELM_DRIVER", "configmap")

	oo, err := h.List(context.Background(), client.BlankNamespace)
	require.NoError(t, err)
	assert.Empty(t, oo)

	assert.Zero(t, a.metaCount())
	a.mx.Lock()
	defer a.mx.Unlock()
	assert.Positive(t, a.cmReqs)
}
