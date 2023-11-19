// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client_test

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
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
			flags:   &genericclioptions.ConfigFlags{KubeConfig: &kubeConfig},
			cluster: "fred",
		},
		"custom": {
			flags:   &genericclioptions.ConfigFlags{KubeConfig: &kubeConfig, ClusterName: &name},
			cluster: "blee",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cfg := client.NewConfig(u.flags)
			ctx, err := cfg.CurrentClusterName()
			assert.Nil(t, err)
			assert.Equal(t, u.cluster, ctx)
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
			err:     errors.New("invalid context `bozo specified"),
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
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := client.NewConfig(&flags)
	err := cfg.SwitchContext("blee")
	assert.Nil(t, err)
	ctx, err := cfg.CurrentContextName()
	assert.Nil(t, err)
	assert.Equal(t, "blee", ctx)
}

func TestConfigClusterNameFromContext(t *testing.T) {
	cluster, kubeConfig := "duh", "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := client.NewConfig(&flags)
	cl, err := cfg.ClusterNameFromContext("blee")
	assert.Nil(t, err)
	assert.Equal(t, "blee", cl)
}

func TestConfigAccess(t *testing.T) {
	cluster, kubeConfig := "duh", "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := client.NewConfig(&flags)
	acc, err := cfg.ConfigAccess()
	assert.Nil(t, err)
	assert.True(t, len(acc.GetDefaultFilename()) > 0)
}

func TestConfigContexts(t *testing.T) {
	cluster, kubeConfig := "duh", "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := client.NewConfig(&flags)
	cc, err := cfg.Contexts()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cc))
}

func TestConfigContextNames(t *testing.T) {
	cluster, kubeConfig := "duh", "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := client.NewConfig(&flags)
	cc, err := cfg.ContextNames()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cc))
}

func TestConfigClusterNames(t *testing.T) {
	cluster, kubeConfig := "duh", "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := client.NewConfig(&flags)
	cc, err := cfg.ClusterNames()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cc))
}

func TestConfigDelContext(t *testing.T) {
	cluster, kubeConfig := "duh", "./testdata/config.1"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := client.NewConfig(&flags)
	err := cfg.DelContext("fred")
	assert.Nil(t, err)
	cc, err := cfg.ContextNames()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(cc))
	assert.Equal(t, "blee", cc[0])
}

func TestConfigRestConfig(t *testing.T) {
	kubeConfig := "./testdata/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
	}

	cfg := client.NewConfig(&flags)
	rc, err := cfg.RESTConfig()
	assert.Nil(t, err)
	assert.Equal(t, "https://localhost:3000", rc.Host)
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

func TestNamespaceNames(t *testing.T) {
	nn := []v1.Namespace{
		{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}},
	}

	nns := client.NamespaceNames(nn)
	assert.Equal(t, 2, len(nns))
	assert.Equal(t, []string{"ns1", "ns2"}, nns)
}
