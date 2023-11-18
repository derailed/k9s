// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

func TestIsReadOnly(t *testing.T) {
	uu := map[string]struct {
		config      string
		read, write bool
		readOnly    bool
	}{
		"writable": {
			config: "k9s.yml",
		},
		"writable_read_override": {
			config:   "k9s.yml",
			read:     true,
			readOnly: true,
		},
		"writable_write_override": {
			config: "k9s.yml",
			write:  true,
		},
		"readonly": {
			config:   "k9s_readonly.yml",
			readOnly: true,
		},
		"readonly_read_override": {
			config:   "k9s_readonly.yml",
			read:     true,
			readOnly: true,
		},
		"readonly_write_override": {
			config: "k9s_readonly.yml",
			write:  true,
		},
		"readonly_both_override": {
			config: "k9s_readonly.yml",
			read:   true,
			write:  true,
		},
	}

	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Nil(t, cfg.Load("testdata/"+u.config))
			cfg.K9s.OverrideReadOnly(u.read)
			cfg.K9s.OverrideWrite(u.write)
			assert.Equal(t, u.readOnly, cfg.K9s.IsReadOnly())
		})
	}
}

func TestK9sValidate(t *testing.T) {
	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)

	mk := NewMockKubeSettings()
	m.When(mk.CurrentContextName()).ThenReturn("ctx1", nil)
	m.When(mk.CurrentClusterName()).ThenReturn("c1", nil)
	m.When(mk.ClusterNames()).ThenReturn(map[string]struct{}{"c1": {}, "c2": {}}, nil)
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"default"})

	c := config.NewK9s()
	c.Validate(mc, mk)

	assert.Equal(t, 2, c.RefreshRate)
	assert.Equal(t, int64(100), c.Logger.TailCount)
	assert.Equal(t, 5000, c.Logger.BufferSize)
	assert.Equal(t, "ctx1", c.CurrentContext)
	assert.Equal(t, "c1", c.CurrentCluster)
	assert.Equal(t, 1, len(c.Clusters))
	assert.Equal(t, config.K9sDefaultScreenDumpDir, c.GetScreenDumpDir())
	_, ok := c.Clusters[c.CurrentCluster]
	assert.True(t, ok)
}

func TestK9sValidateBlank(t *testing.T) {
	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)

	mk := NewMockKubeSettings()
	m.When(mk.CurrentContextName()).ThenReturn("ctx1", nil)
	m.When(mk.CurrentClusterName()).ThenReturn("c1", nil)
	m.When(mk.ClusterNames()).ThenReturn(map[string]struct{}{"c1": {}, "c2": {}}, nil)
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"default"})

	var c config.K9s
	c.Validate(mc, mk)

	assert.Equal(t, 2, c.RefreshRate)
	assert.Equal(t, int64(100), c.Logger.TailCount)
	assert.Equal(t, 5000, c.Logger.BufferSize)
	assert.Equal(t, "ctx1", c.CurrentContext)
	assert.Equal(t, "c1", c.CurrentCluster)
	assert.Equal(t, 1, len(c.Clusters))
	_, ok := c.Clusters[c.CurrentCluster]
	assert.True(t, ok)
}

func TestK9sActiveClusterZero(t *testing.T) {
	c := config.NewK9s()
	c.CurrentCluster = "fred"
	cl := c.ActiveCluster()
	assert.NotNil(t, cl)
	assert.Equal(t, "default", cl.Namespace.Active)
	assert.Equal(t, 1, len(cl.Namespace.Favorites))
}

func TestK9sActiveClusterBlank(t *testing.T) {
	var c config.K9s
	cl := c.ActiveCluster()
	assert.Equal(t, config.NewCluster(), cl)
}

func TestK9sActiveCluster(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.Nil(t, cfg.Load("testdata/k9s.yml"))

	cl := cfg.K9s.ActiveCluster()
	assert.NotNil(t, cl)
	assert.Equal(t, "kube-system", cl.Namespace.Active)
	assert.Equal(t, 5, len(cl.Namespace.Favorites))
}

func TestGetScreenDumpDir(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.Nil(t, cfg.Load("testdata/k9s.yml"))

	assert.Equal(t, "/tmp", cfg.K9s.GetScreenDumpDir())
}

func TestGetScreenDumpDirOverride(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	cfg.K9s.OverrideScreenDumpDir("/override")

	assert.Equal(t, "/override", cfg.K9s.GetScreenDumpDir())
}

func TestGetScreenDumpDirOverrideEmpty(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	cfg.K9s.OverrideScreenDumpDir("")

	assert.Equal(t, "/tmp", cfg.K9s.GetScreenDumpDir())
}

func TestGetScreenDumpDirEmpty(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.Nil(t, cfg.Load("testdata/k9s1.yml"))
	cfg.K9s.OverrideScreenDumpDir("")

	assert.Equal(t, config.K9sDefaultScreenDumpDir, cfg.K9s.GetScreenDumpDir())
}
