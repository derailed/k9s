// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/mock"
	m "github.com/petergtz/pegomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestConfigSave(t *testing.T) {
	config.AppConfigFile = "/tmp/k9s-test/k9s.yaml"
	sd := "/tmp/k9s-test/screen-dumps"
	cl, ct := "cl-1", "ct-1-1"
	_ = os.RemoveAll(("/tmp/k9s-test"))

	uu := map[string]struct {
		ct       string
		flags    *genericclioptions.ConfigFlags
		k9sFlags *config.Flags
	}{
		"happy": {
			ct: "ct-1-1",
			flags: &genericclioptions.ConfigFlags{
				ClusterName: &cl,
				Context:     &ct,
			},
			k9sFlags: &config.Flags{
				ScreenDumpDir: &sd,
			},
		},
	}

	for k := range uu {
		xdg.Reload()
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := mock.NewMockConfig()
			_, err := c.K9s.ActivateContext(u.ct)
			assert.NoError(t, err)
			if u.flags != nil {
				c.K9s.Override(u.k9sFlags)
				assert.NoError(t, c.Refine(u.flags, u.k9sFlags, client.NewConfig(u.flags)))
			}
			assert.NoError(t, c.Save(true))
			bb, err := os.ReadFile(config.AppConfigFile)
			assert.NoError(t, err)
			ee, err := os.ReadFile("testdata/configs/default.yaml")
			assert.NoError(t, err)
			assert.Equal(t, string(ee), string(bb))
		})
	}
}

func TestSetActiveView(t *testing.T) {
	var (
		cfgFile = "testdata/kubes/test.yaml"
		view    = "dp"
	)

	uu := map[string]struct {
		ct       string
		flags    *genericclioptions.ConfigFlags
		k9sFlags *config.Flags
		view     string
		e        string
	}{
		"empty": {
			view: data.DefaultView,
			e:    data.DefaultView,
		},
		"not-exists": {
			ct:   "fred",
			view: data.DefaultView,
			e:    data.DefaultView,
		},
		"happy": {
			ct:   "ct-1-1",
			view: "xray",
			e:    "xray",
		},
		"cli-override": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
			},
			k9sFlags: &config.Flags{
				Command: &view,
			},
			ct:   "ct-1-1",
			view: "xray",
			e:    "dp",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := mock.NewMockConfig()
			_, _ = c.K9s.ActivateContext(u.ct)
			if u.flags != nil {
				assert.NoError(t, c.Refine(u.flags, nil, client.NewConfig(u.flags)))
				c.K9s.Override(u.k9sFlags)
			}
			c.SetActiveView(u.view)
			assert.Equal(t, u.e, c.ActiveView())
		})
	}
}

func TestActiveContextName(t *testing.T) {
	var (
		cfgFile = "testdata/kubes/test.yaml"
		ct2     = "ct-1-2"
	)

	uu := map[string]struct {
		flags    *genericclioptions.ConfigFlags
		k9sFlags *config.Flags
		ct       string
		e        string
	}{
		"empty": {},
		"happy": {
			ct: "ct-1-1",
			e:  "ct-1-1",
		},
		"cli-override": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
				Context:    &ct2,
			},
			k9sFlags: &config.Flags{},
			ct:       "ct-1-1",
			e:        "ct-1-2",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := mock.NewMockConfig()
			_, _ = c.K9s.ActivateContext(u.ct)
			if u.flags != nil {
				assert.NoError(t, c.Refine(u.flags, nil, client.NewConfig(u.flags)))
				c.K9s.Override(u.k9sFlags)
			}
			assert.Equal(t, u.e, c.ActiveContextName())
		})
	}
}

func TestActiveView(t *testing.T) {
	var (
		cfgFile = "testdata/kubes/test.yaml"
		view    = "dp"
	)

	uu := map[string]struct {
		ct       string
		flags    *genericclioptions.ConfigFlags
		k9sFlags *config.Flags
		e        string
	}{
		"empty": {
			e: data.DefaultView,
		},
		"not-exists": {
			ct: "fred",
			e:  data.DefaultView,
		},
		"happy": {
			ct: "ct-1-1",
			e:  data.DefaultView,
		},
		"cli-override": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
			},
			k9sFlags: &config.Flags{
				Command: &view,
			},
			e: "dp",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := mock.NewMockConfig()
			_, _ = c.K9s.ActivateContext(u.ct)
			if u.flags != nil {
				assert.NoError(t, c.Refine(u.flags, nil, client.NewConfig(u.flags)))
				c.K9s.Override(u.k9sFlags)
			}
			assert.Equal(t, u.e, c.ActiveView())
		})
	}
}

func TestFavNamespaces(t *testing.T) {
	uu := map[string]struct {
		ct string
		e  []string
	}{
		"empty": {},
		"not-exists": {
			ct: "fred",
		},
		"happy": {
			ct: "ct-1-1",
			e:  []string{client.DefaultNamespace},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := mock.NewMockConfig()
			_, _ = c.K9s.ActivateContext(u.ct)
			assert.Equal(t, u.e, c.FavNamespaces())
		})
	}
}

func TestContextAliasesPath(t *testing.T) {
	uu := map[string]struct {
		ct string
		e  string
	}{
		"empty": {},
		"not-exists": {
			ct: "fred",
		},
		"happy": {
			ct: "ct-1-1",
			e:  "/tmp/test/cl-1/ct-1-1/aliases.yaml",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := mock.NewMockConfig()
			_, _ = c.K9s.ActivateContext(u.ct)
			assert.Equal(t, u.e, c.ContextAliasesPath())
		})
	}
}

func TestContextPluginsPath(t *testing.T) {
	uu := map[string]struct {
		ct, e string
		err   error
	}{
		"empty": {
			err: errors.New(`no context found for: ""`),
		},
		"happy": {
			ct: "ct-1-1",
			e:  "/tmp/test/cl-1/ct-1-1/plugins.yaml",
		},
		"not-exists": {
			ct:  "fred",
			err: errors.New(`no context found for: "fred"`),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := mock.NewMockConfig()
			_, _ = c.K9s.ActivateContext(u.ct)
			s, err := c.ContextPluginsPath()
			if err != nil {
				assert.Equal(t, u.err, err)
			}
			assert.Equal(t, u.e, s)
		})
	}
}

func TestConfigLoader(t *testing.T) {
	uu := map[string]struct {
		f   string
		err string
	}{
		"happy": {
			f: "testdata/configs/k9s.yaml",
		},
		"toast": {
			f: "testdata/configs/k9s_toast.yaml",
			err: `k9s config file "testdata/configs/k9s_toast.yaml" load failed:
Additional property disablePodCounts is not allowed
Additional property shellPods is not allowed
Invalid type. Expected: boolean, given: string`,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cfg := config.NewConfig(nil)
			if err := cfg.Load(u.f, true); err != nil {
				assert.Equal(t, u.err, err.Error())
			}
		})
	}
}

func TestConfigSetCurrentContext(t *testing.T) {
	uu := map[string]struct {
		cl, ct string
		err    string
	}{
		"happy": {
			ct: "ct-1-2",
			cl: "cl-1",
		},
		"toast": {
			ct:  "fred",
			cl:  "cl-1",
			err: `set current context failed. no context found for: "fred"`,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cfg := mock.NewMockConfig()
			ct, err := cfg.SetCurrentContext(u.ct)
			if err != nil {
				assert.Equal(t, u.err, err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, u.cl, ct.ClusterName)
		})
	}
}

func TestConfigCurrentContext(t *testing.T) {
	var (
		cfgFile = "testdata/kubes/test.yaml"
		ct2     = "ct-1-2"
	)

	uu := map[string]struct {
		flags     *genericclioptions.ConfigFlags
		err       error
		context   string
		cluster   string
		namespace string
	}{
		"override-context": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
				Context:    &ct2,
			},
			cluster:   "cl-1",
			context:   "ct-1-2",
			namespace: "ns-2",
		},
		"use-current-context": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
			},
			cluster:   "cl-1",
			context:   "ct-1-1",
			namespace: client.DefaultNamespace,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cfg := mock.NewMockConfig()

			err := cfg.Refine(u.flags, nil, client.NewConfig(u.flags))
			assert.NoError(t, err)
			ct, err := cfg.CurrentContext()
			assert.NoError(t, err)
			assert.Equal(t, u.cluster, ct.ClusterName)
			assert.Equal(t, u.namespace, ct.Namespace.Active)
		})
	}
}

func TestConfigRefine(t *testing.T) {
	var (
		cfgFile       = "testdata/kubes/test.yaml"
		cl1           = "cl-1"
		ct2           = "ct-1-2"
		ns1, ns2, nsx = "ns-1", "ns-2", "ns-x"
		true          = true
	)

	uu := map[string]struct {
		flags     *genericclioptions.ConfigFlags
		k9sFlags  *config.Flags
		err       string
		context   string
		cluster   string
		namespace string
	}{
		"no-override": {
			namespace: "default",
		},
		"override-cluster": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig:  &cfgFile,
				ClusterName: &cl1,
			},
			cluster:   "cl-1",
			context:   "ct-1-1",
			namespace: client.DefaultNamespace,
		},
		"override-cluster-context": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig:  &cfgFile,
				ClusterName: &cl1,
				Context:     &ct2,
			},
			cluster:   "cl-1",
			context:   "ct-1-2",
			namespace: "ns-2",
		},
		"override-bad-cluster": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig:  &cfgFile,
				ClusterName: &ns1,
			},
			cluster:   "cl-1",
			context:   "ct-1-1",
			namespace: client.DefaultNamespace,
		},
		"override-ns": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
				Namespace:  &ns2,
			},
			cluster:   "cl-1",
			context:   "ct-1-1",
			namespace: "ns-2",
		},
		"all-ns": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
				Namespace:  &ns2,
			},
			k9sFlags: &config.Flags{
				AllNamespaces: &true,
			},
			cluster:   "cl-1",
			context:   "ct-1-1",
			namespace: client.NamespaceAll,
		},

		"override-bad-ns": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
				Namespace:  &nsx,
			},
			cluster:   "cl-1",
			context:   "ct-1-1",
			namespace: "ns-x",
		},
		"override-context": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
				Context:    &ct2,
			},
			cluster:   "cl-1",
			context:   "ct-1-2",
			namespace: "ns-2",
		},
		"override-bad-context": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
				Context:    &ns1,
			},
			err: `k8sflags. unable to activate context "ns-1": no context found for: "ns-1"`,
		},
		"use-current-context": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &cfgFile,
			},
			cluster:   "cl-1",
			context:   "ct-1-1",
			namespace: client.DefaultNamespace,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cfg := mock.NewMockConfig()

			err := cfg.Refine(u.flags, u.k9sFlags, client.NewConfig(u.flags))
			if err != nil {
				assert.Equal(t, u.err, err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, u.context, cfg.K9s.ActiveContextName())
				assert.Equal(t, u.namespace, cfg.ActiveNamespace())
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := mock.NewMockConfig()
	cfg.SetConnection(mock.NewMockConnection())

	assert.Nil(t, cfg.Load("testdata/configs/k9s.yaml", true))
	cfg.Validate()
}

func TestConfigLoad(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.Nil(t, cfg.Load("testdata/configs/k9s.yaml", true))
	assert.Equal(t, 2, cfg.K9s.RefreshRate)
	assert.Equal(t, int64(200), cfg.K9s.Logger.TailCount)
	assert.Equal(t, 2000, cfg.K9s.Logger.BufferSize)
}

func TestConfigLoadCrap(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.NotNil(t, cfg.Load("testdata/configs/k9s_not_there.yaml", true))
}

func TestConfigSaveFile(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.Nil(t, cfg.Load("testdata/configs/k9s.yaml", true))

	cfg.K9s.RefreshRate = 100
	cfg.K9s.ReadOnly = true
	cfg.K9s.Logger.TailCount = 500
	cfg.K9s.Logger.BufferSize = 800
	cfg.Validate()

	path := filepath.Join("/tmp", "k9s.yaml")
	assert.NoError(t, cfg.SaveFile(path))
	raw, err := os.ReadFile(path)
	assert.Nil(t, err)
	ee, err := os.ReadFile("testdata/configs/expected.yaml")
	assert.Nil(t, err)
	assert.Equal(t, string(ee), string(raw))
}

func TestConfigReset(t *testing.T) {
	cfg := mock.NewMockConfig()
	assert.Nil(t, cfg.Load("testdata/configs/k9s.yaml", true))
	cfg.Reset()
	cfg.Validate()

	path := filepath.Join("/tmp", "k9s.yaml")
	assert.NoError(t, cfg.SaveFile(path))

	bb, err := os.ReadFile(path)
	assert.Nil(t, err)
	ee, err := os.ReadFile("testdata/configs/k9s.yaml")
	assert.Nil(t, err)
	assert.Equal(t, string(ee), string(bb))
}

// Helpers...

func TestSetup(t *testing.T) {
	m.RegisterMockTestingT(t)
	m.RegisterMockFailHandler(func(m string, i ...int) {
		fmt.Println("Boom!", m, i)
	})
}
