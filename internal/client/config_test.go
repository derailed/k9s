// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client_test

import (
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var kubeConfig = "./testdata/config"

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func TestCallTimeout(t *testing.T) {
	uu := map[string]struct {
		t string
		e time.Duration
	}{
		"custom": {
			t: "1m",
			e: 1 * time.Minute,
		},
		"default": {
			e: 15 * time.Second,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			flags := genericclioptions.NewConfigFlags(false)
			flags.Timeout = &u.t
			cfg := client.NewConfig(flags)
			assert.Equal(t, u.e, cfg.CallTimeout())
		})
	}
}

func TestConfigCurrentContext(t *testing.T) {
	uu := map[string]struct {
		context string
		e       string
	}{
		"default": {
			e: "fred",
		},
		"custom": {
			context: "blee",
			e:       "blee",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			flags := genericclioptions.NewConfigFlags(false)
			flags.KubeConfig = &kubeConfig
			if u.context != "" {
				flags.Context = &u.context
			}
			cfg := client.NewConfig(flags)
			ctx, err := cfg.CurrentContextName()
			require.NoError(t, err)
			assert.Equal(t, u.e, ctx)
		})
	}
}

func TestConfigCurrentCluster(t *testing.T) {
	name := "blee"
	uu := map[string]struct {
		flags   *genericclioptions.ConfigFlags
		cluster string
	}{
		"default": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &kubeConfig,
			},
			cluster: "zorg",
		},
		"custom": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig: &kubeConfig,
				Context:    &name,
			},
			cluster: "blee",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cfg := client.NewConfig(u.flags)
			ct, err := cfg.CurrentClusterName()
			require.NoError(t, err)
			assert.Equal(t, u.cluster, ct)
		})
	}
}

func TestConfigCurrentUser(t *testing.T) {
	name := "blee"
	uu := map[string]struct {
		flags *genericclioptions.ConfigFlags
		user  string
	}{
		"default": {
			flags: &genericclioptions.ConfigFlags{KubeConfig: &kubeConfig},
			user:  "fred",
		},
		"custom": {
			flags: &genericclioptions.ConfigFlags{KubeConfig: &kubeConfig, AuthInfoName: &name},
			user:  "blee",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cfg := client.NewConfig(u.flags)
			ctx, err := cfg.CurrentUserName()
			require.NoError(t, err)
			assert.Equal(t, u.user, ctx)
		})
	}
}

func TestConfigCurrentNamespace(t *testing.T) {
	bleeNS, bleeCTX := "blee", "blee"
	uu := map[string]struct {
		flags     *genericclioptions.ConfigFlags
		namespace string
	}{
		"default": {
			flags:     &genericclioptions.ConfigFlags{KubeConfig: &kubeConfig},
			namespace: "",
		},
		"withContext": {
			flags:     &genericclioptions.ConfigFlags{KubeConfig: &kubeConfig, Context: &bleeCTX},
			namespace: "zorg",
		},
		"withNS": {
			flags:     &genericclioptions.ConfigFlags{KubeConfig: &kubeConfig, Namespace: &bleeNS},
			namespace: "blee",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cfg := client.NewConfig(u.flags)
			ns, err := cfg.CurrentNamespaceName()
			if ns != "" {
				require.NoError(t, err)
			}
			assert.Equal(t, u.namespace, ns)
		})
	}
}

func TestConfigGetContext(t *testing.T) {
	uu := map[string]struct {
		cluster string
		err     error
	}{
		"default": {
			cluster: "blee",
		},
		"custom": {
			cluster: "bozo",
			err:     errors.New(`getcontext - invalid context specified: "bozo"`),
		},
	}

	flags := &genericclioptions.ConfigFlags{KubeConfig: &kubeConfig}
	cfg := client.NewConfig(flags)
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			ctx, err := cfg.GetContext(u.cluster)
			if err != nil {
				assert.Equal(t, u.err, err)
			} else {
				assert.NotNil(t, ctx)
				assert.Equal(t, u.cluster, ctx.Cluster)
			}
		})
	}
}

func TestConfigSwitchContext(t *testing.T) {
	cluster := "duh"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
		Context:    &cluster,
	}

	cfg := client.NewConfig(&flags)
	err := cfg.SwitchContext("blee")
	require.NoError(t, err)
	ctx, err := cfg.CurrentContextName()
	require.NoError(t, err)
	assert.Equal(t, "blee", ctx)
}

func TestConfigAccess(t *testing.T) {
	context := "duh"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
		Context:    &context,
	}

	cfg := client.NewConfig(&flags)
	acc, err := cfg.ConfigAccess()
	require.NoError(t, err)
	assert.NotEmpty(t, acc.GetDefaultFilename())
}

func TestConfigContextNames(t *testing.T) {
	cluster := "duh"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
		Context:    &cluster,
	}

	cfg := client.NewConfig(&flags)
	cc, err := cfg.ContextNames()
	require.NoError(t, err)
	assert.Len(t, cc, 3)
}

func TestConfigContexts(t *testing.T) {
	context := "duh"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
		Context:    &context,
	}

	cfg := client.NewConfig(&flags)
	cc, err := cfg.Contexts()
	require.NoError(t, err)
	assert.Len(t, cc, 3)
}

func TestConfigDelContext(t *testing.T) {
	require.NoError(t, cp("./testdata/config.2", "./testdata/config.1"))

	context, kubeCfg := "duh", "./testdata/config.1"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeCfg,
		Context:    &context,
	}

	cfg := client.NewConfig(&flags)
	err := cfg.DelContext("fred")
	require.NoError(t, err)

	cc, err := cfg.ContextNames()
	require.NoError(t, err)
	assert.Len(t, cc, 1)
	_, ok := cc["blee"]
	assert.True(t, ok)
}

func TestConfigRestConfig(t *testing.T) {
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
	}

	cfg := client.NewConfig(&flags)
	rc, err := cfg.RESTConfig()
	require.NoError(t, err)
	assert.Equal(t, "https://localhost:3002", rc.Host)
}

func TestConfigBadConfig(t *testing.T) {
	kubeConfig := "./testdata/bork_config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
	}

	cfg := client.NewConfig(&flags)
	_, err := cfg.RESTConfig()
	assert.Error(t, err)
}

// Helpers...

func cp(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0600)
}
