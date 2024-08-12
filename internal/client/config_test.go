// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
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
	var kubeConfig = "./testdata/config"

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
			assert.Nil(t, err)
			assert.Equal(t, u.e, ctx)
		})
	}
}

func TestConfigCurrentCluster(t *testing.T) {
	name, kubeConfig := "blee", "./testdata/config"
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
			assert.Nil(t, err)
			assert.Equal(t, u.cluster, ct)
		})
	}
}

func TestConfigCurrentUser(t *testing.T) {
	name, kubeConfig := "blee", "./testdata/config"
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
			assert.Nil(t, err)
			assert.Equal(t, u.user, ctx)
		})
	}
}

func TestConfigCurrentNamespace(t *testing.T) {
	kubeConfig := "./testdata/config"
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
				assert.Nil(t, err)
			}
			assert.Equal(t, u.namespace, ns)
		})
	}
}

func TestConfigGetContext(t *testing.T) {
	kubeConfig := "./testdata/config"
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
	cluster, kubeConfig := "duh", "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
		Context:    &cluster,
	}

	cfg := client.NewConfig(&flags)
	err := cfg.SwitchContext("blee")
	assert.Nil(t, err)
	ctx, err := cfg.CurrentContextName()
	assert.Nil(t, err)
	assert.Equal(t, "blee", ctx)
}

func TestConfigAccess(t *testing.T) {
	context, kubeConfig := "duh", "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
		Context:    &context,
	}

	cfg := client.NewConfig(&flags)
	acc, err := cfg.ConfigAccess()
	assert.Nil(t, err)
	assert.True(t, len(acc.GetDefaultFilename()) > 0)
}

func TestConfigContextNames(t *testing.T) {
	cluster, kubeConfig := "duh", "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
		Context:    &cluster,
	}

	cfg := client.NewConfig(&flags)
	cc, err := cfg.ContextNames()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cc))
}

func TestConfigContexts(t *testing.T) {
	context, kubeConfig := "duh", "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
		Context:    &context,
	}

	cfg := client.NewConfig(&flags)
	cc, err := cfg.Contexts()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cc))
}

func TestConfigDelContext(t *testing.T) {
	assert.NoError(t, cp("./testdata/config.2", "./testdata/config.1"))

	context, kubeConfig := "duh", "./testdata/config.1"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
		Context:    &context,
	}

	cfg := client.NewConfig(&flags)
	err := cfg.DelContext("fred")
	assert.NoError(t, err)

	cc, err := cfg.ContextNames()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(cc))
	_, ok := cc["blee"]
	assert.True(t, ok)
}

func TestConfigRestConfig(t *testing.T) {
	kubeConfig := "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
	}

	cfg := client.NewConfig(&flags)
	rc, err := cfg.RESTConfig()
	assert.Nil(t, err)
	assert.Equal(t, "https://localhost:3002", rc.Host)
}

func TestConfigBadConfig(t *testing.T) {
	kubeConfig := "./testdata/bork_config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
	}

	cfg := client.NewConfig(&flags)
	_, err := cfg.RESTConfig()
	assert.NotNil(t, err)
}

// Helpers...

func cp(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0600)
}
