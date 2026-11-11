// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// kubeconfigFile writes kubeconfig revisions and makes each one look changed
// on disk regardless of the file system timestamp granularity.
type kubeconfigFile struct {
	t     *testing.T
	path  string
	mtime time.Time
}

func newKubeconfigFile(t *testing.T, cfg *api.Config) *kubeconfigFile {
	t.Helper()

	f := kubeconfigFile{
		t:     t,
		path:  filepath.Join(t.TempDir(), "config"),
		mtime: time.Now().Add(-time.Hour),
	}
	f.write(cfg)

	return &f
}

func (f *kubeconfigFile) write(cfg *api.Config) {
	f.t.Helper()

	require.NoError(f.t, clientcmd.WriteToFile(*cfg, f.path))
	f.touch()
}

func (f *kubeconfigFile) writeRaw(s string) {
	f.t.Helper()

	require.NoError(f.t, os.WriteFile(f.path, []byte(s), 0o600))
	f.touch()
}

func (f *kubeconfigFile) touch() {
	f.t.Helper()

	f.mtime = f.mtime.Add(time.Second)
	require.NoError(f.t, os.Chtimes(f.path, f.mtime, f.mtime))
}

func (f *kubeconfigFile) config() *client.Config {
	flags := genericclioptions.NewConfigFlags(client.UsePersistentConfig)
	flags.KubeConfig = &f.path

	return client.NewConfig(flags)
}

// testKubeconfig has contexts c1 and c2 backed by their own cluster and user.
func testKubeconfig(current string) *api.Config {
	cfg := api.NewConfig()
	for _, n := range []string{"c1", "c2"} {
		cfg.Clusters[n] = &api.Cluster{Server: "https://" + n + ".example.com"}
		cfg.AuthInfos[n] = &api.AuthInfo{Token: n + "-token"}
		cfg.Contexts[n] = &api.Context{Cluster: n, AuthInfo: n}
	}
	cfg.CurrentContext = current

	return cfg
}

func contextNames(t *testing.T, cfg *client.Config) []string {
	t.Helper()

	cc, err := cfg.ContextNames()
	require.NoError(t, err)
	nn := make([]string, 0, len(cc))
	for n := range cc {
		nn = append(nn, n)
	}

	return nn
}

func bearerToken(t *testing.T, cfg *client.Config) string {
	t.Helper()

	rc, err := cfg.RESTConfig()
	require.NoError(t, err)

	return rc.BearerToken
}

func TestKubeconfigReloadListsNewContexts(t *testing.T) {
	kc := testKubeconfig("c1")
	f := newKubeconfigFile(t, kc)
	cfg := f.config()
	assert.ElementsMatch(t, []string{"c1", "c2"}, contextNames(t, cfg))

	kc.Contexts["c3"] = &api.Context{Cluster: "c1", AuthInfo: "c1"}
	f.write(kc)

	assert.ElementsMatch(t, []string{"c1", "c2", "c3"}, contextNames(t, cfg))
	ct, err := cfg.GetContext("c3")
	require.NoError(t, err)
	assert.Equal(t, "c1", ct.Cluster)
	assert.Equal(t, uint64(0), cfg.ActiveRevision())
}

func TestKubeconfigReloadKeepsActiveContextPinned(t *testing.T) {
	kc := testKubeconfig("c1")
	f := newKubeconfigFile(t, kc)
	cfg := f.config()
	n, err := cfg.CurrentContextName()
	require.NoError(t, err)
	assert.Equal(t, "c1", n)

	kc.CurrentContext = "c2"
	f.write(kc)

	n, err = cfg.CurrentContextName()
	require.NoError(t, err)
	assert.Equal(t, "c1", n)
	cl, err := cfg.CurrentClusterName()
	require.NoError(t, err)
	assert.Equal(t, "c1", cl)
	u, err := cfg.CurrentUserName()
	require.NoError(t, err)
	assert.Equal(t, "c1", u)
	assert.Equal(t, "c1-token", bearerToken(t, cfg))
	assert.Empty(t, cfg.DetachedReason())
}

func TestKubeconfigReloadAppliesActiveSettings(t *testing.T) {
	uu := map[string]struct {
		edit func(*api.Config)
		rev  uint64
		tok  string
	}{
		"active-token": {
			edit: func(c *api.Config) { c.AuthInfos["c1"].Token = "fresh" },
			rev:  1,
			tok:  "fresh",
		},
		"active-server": {
			edit: func(c *api.Config) { c.Clusters["c1"].Server = "https://moved.example.com" },
			rev:  1,
			tok:  "c1-token",
		},
		"active-user-swap": {
			edit: func(c *api.Config) { c.Contexts["c1"].AuthInfo = "c2" },
			rev:  1,
			tok:  "c2-token",
		},
		"other-context-token": {
			edit: func(c *api.Config) { c.AuthInfos["c2"].Token = "fresh" },
			tok:  "c1-token",
		},
		"active-namespace": {
			edit: func(c *api.Config) { c.Contexts["c1"].Namespace = "fred" },
			tok:  "c1-token",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			kc := testKubeconfig("c1")
			f := newKubeconfigFile(t, kc)
			cfg := f.config()
			assert.Equal(t, "c1-token", bearerToken(t, cfg))

			u.edit(kc)
			f.write(kc)

			assert.Equal(t, u.rev, cfg.ActiveRevision())
			assert.Equal(t, u.tok, bearerToken(t, cfg))
			assert.Empty(t, cfg.DetachedReason())
		})
	}
}

func TestKubeconfigReloadDetachesActiveContext(t *testing.T) {
	uu := map[string]struct {
		edit   func(*api.Config)
		reason string
	}{
		"removed": {
			edit:   func(c *api.Config) { delete(c.Contexts, "c1") },
			reason: `Active context "c1" was removed from the kubeconfig`,
		},
		"renamed": {
			edit: func(c *api.Config) {
				c.Contexts["c1-renamed"] = c.Contexts["c1"]
				delete(c.Contexts, "c1")
			},
			reason: `Active context "c1" was removed from the kubeconfig`,
		},
		"cluster-swap": {
			edit:   func(c *api.Config) { c.Contexts["c1"].Cluster = "c2" },
			reason: `Active context "c1" now points at cluster "c2" instead of "c1"`,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			kc := testKubeconfig("c1")
			f := newKubeconfigFile(t, kc)
			cfg := f.config()
			assert.Equal(t, "c1-token", bearerToken(t, cfg))

			u.edit(kc)
			kc.AuthInfos["c1"].Token = "fresh"
			f.write(kc)

			assert.Equal(t, u.reason, cfg.DetachedReason())
			assert.Equal(t, uint64(0), cfg.ActiveRevision())
			assert.Equal(t, "c1-token", bearerToken(t, cfg))
			n, err := cfg.CurrentContextName()
			require.NoError(t, err)
			assert.Equal(t, "c1", n)
			cl, err := cfg.CurrentClusterName()
			require.NoError(t, err)
			assert.Equal(t, "c1", cl)
			ct, err := cfg.GetContext("c1")
			require.NoError(t, err)
			assert.Equal(t, "c1", ct.Cluster)

			f.write(testKubeconfig("c2"))

			assert.Empty(t, cfg.DetachedReason())
			assert.Equal(t, uint64(0), cfg.ActiveRevision())
			n, err = cfg.CurrentContextName()
			require.NoError(t, err)
			assert.Equal(t, "c1", n)
		})
	}
}

func TestKubeconfigReloadOnlyDetachedActiveContextResolves(t *testing.T) {
	kc := testKubeconfig("c1")
	f := newKubeconfigFile(t, kc)
	cfg := f.config()
	assert.ElementsMatch(t, []string{"c1", "c2"}, contextNames(t, cfg))

	delete(kc.Contexts, "c1")
	delete(kc.Contexts, "c2")
	f.write(kc)

	assert.Empty(t, contextNames(t, cfg))
	_, err := cfg.GetContext("c1")
	require.NoError(t, err)
	_, err = cfg.GetContext("c2")
	require.Error(t, err)
}

func TestKubeconfigReloadIgnoresBrokenFile(t *testing.T) {
	kc := testKubeconfig("c1")
	f := newKubeconfigFile(t, kc)
	cfg := f.config()
	assert.ElementsMatch(t, []string{"c1", "c2"}, contextNames(t, cfg))

	f.writeRaw("apiVersion: v1\nkind: Config\ncontexts: [")

	assert.ElementsMatch(t, []string{"c1", "c2"}, contextNames(t, cfg))
	assert.Equal(t, "c1-token", bearerToken(t, cfg))

	kc.AuthInfos["c1"].Token = "fresh"
	f.write(kc)

	assert.Equal(t, uint64(1), cfg.ActiveRevision())
	assert.Equal(t, "fresh", bearerToken(t, cfg))
}

func TestKubeconfigReloadWithoutActiveContext(t *testing.T) {
	kc := testKubeconfig("")
	f := newKubeconfigFile(t, kc)
	cfg := f.config()
	n, err := cfg.CurrentContextName()
	require.NoError(t, err)
	assert.Empty(t, n)

	kc.CurrentContext = "c1"
	f.write(kc)

	n, err = cfg.CurrentContextName()
	require.NoError(t, err)
	assert.Empty(t, n)
	assert.ElementsMatch(t, []string{"c1", "c2"}, contextNames(t, cfg))
}

func TestKubeconfigReloadMergedFiles(t *testing.T) {
	base := testKubeconfig("c1")
	f1 := newKubeconfigFile(t, base)
	extra := api.NewConfig()
	extra.Clusters["x"] = &api.Cluster{Server: "https://x.example.com"}
	extra.AuthInfos["x"] = &api.AuthInfo{Token: "x-token"}
	f2 := newKubeconfigFile(t, extra)
	t.Setenv(clientcmd.RecommendedConfigPathEnvVar, f1.path+string(filepath.ListSeparator)+f2.path)

	cfg := client.NewConfig(genericclioptions.NewConfigFlags(client.UsePersistentConfig))
	assert.ElementsMatch(t, []string{"c1", "c2"}, contextNames(t, cfg))

	extra.Contexts["x"] = &api.Context{Cluster: "x", AuthInfo: "x"}
	f2.write(extra)

	assert.ElementsMatch(t, []string{"c1", "c2", "x"}, contextNames(t, cfg))
	n, err := cfg.CurrentContextName()
	require.NoError(t, err)
	assert.Equal(t, "c1", n)
}

func TestKubeconfigReloadAfterInAppEdits(t *testing.T) {
	f := newKubeconfigFile(t, testKubeconfig("c1"))
	cfg := f.config()

	require.NoError(t, cfg.DelContext("c2"))
	assert.ElementsMatch(t, []string{"c1"}, contextNames(t, cfg))

	require.NoError(t, cfg.RenameContext("c1", "c1-renamed"))
	assert.ElementsMatch(t, []string{"c1-renamed"}, contextNames(t, cfg))
	n, err := cfg.CurrentContextName()
	require.NoError(t, err)
	assert.Equal(t, "c1-renamed", n)
	assert.Empty(t, cfg.DetachedReason())
}

func TestKubeconfigReloadSwitchContextPicksUpEdits(t *testing.T) {
	kc := testKubeconfig("c1")
	f := newKubeconfigFile(t, kc)
	cfg := f.config()
	assert.Equal(t, "c1-token", bearerToken(t, cfg))

	require.NoError(t, cfg.SwitchContext("c2"))
	assert.Equal(t, "c2-token", bearerToken(t, cfg))

	kc.AuthInfos["c2"].Token = "fresh"
	f.write(kc)

	assert.Equal(t, uint64(1), cfg.ActiveRevision())
	assert.Equal(t, "fresh", bearerToken(t, cfg))
}
