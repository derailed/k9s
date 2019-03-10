package k8s_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestConfigCurrentContext(t *testing.T) {
	name, kubeConfig := "blee", "./assets/config"
	uu := []struct {
		flags   *genericclioptions.ConfigFlags
		context string
	}{
		{&genericclioptions.ConfigFlags{KubeConfig: &kubeConfig}, "fred"},
		{&genericclioptions.ConfigFlags{KubeConfig: &kubeConfig, Context: &name}, "blee"},
	}

	for _, u := range uu {
		cfg := k8s.NewConfig(u.flags)
		ctx, err := cfg.CurrentContextName()
		assert.Nil(t, err)
		assert.Equal(t, u.context, ctx)
	}
}

func TestConfigCurrentCluster(t *testing.T) {
	name, kubeConfig := "blee", "./assets/config"
	uu := []struct {
		flags   *genericclioptions.ConfigFlags
		cluster string
	}{
		{&genericclioptions.ConfigFlags{KubeConfig: &kubeConfig}, "fred"},
		{&genericclioptions.ConfigFlags{KubeConfig: &kubeConfig, ClusterName: &name}, "blee"},
	}

	for _, u := range uu {
		cfg := k8s.NewConfig(u.flags)
		ctx, err := cfg.CurrentClusterName()
		assert.Nil(t, err)
		assert.Equal(t, u.cluster, ctx)
	}
}

func TestConfigCurrentUser(t *testing.T) {
	name, kubeConfig := "blee", "./assets/config"
	uu := []struct {
		flags *genericclioptions.ConfigFlags
		user  string
	}{
		{&genericclioptions.ConfigFlags{KubeConfig: &kubeConfig}, "fred"},
		{&genericclioptions.ConfigFlags{KubeConfig: &kubeConfig, AuthInfoName: &name}, "blee"},
	}

	for _, u := range uu {
		cfg := k8s.NewConfig(u.flags)
		ctx, err := cfg.CurrentUserName()
		assert.Nil(t, err)
		assert.Equal(t, u.user, ctx)
	}
}

func TestConfigCurrentNamespace(t *testing.T) {
	name, kubeConfig := "blee", "./assets/config"
	uu := []struct {
		flags     *genericclioptions.ConfigFlags
		namespace string
		err       error
	}{
		{&genericclioptions.ConfigFlags{KubeConfig: &kubeConfig}, "", fmt.Errorf("No active namespace specified")},
		{&genericclioptions.ConfigFlags{KubeConfig: &kubeConfig, Namespace: &name}, "blee", nil},
	}

	for _, u := range uu {
		cfg := k8s.NewConfig(u.flags)
		ns, err := cfg.CurrentNamespaceName()
		assert.Equal(t, u.err, err)
		assert.Equal(t, u.namespace, ns)
	}
}

func TestConfigGetContext(t *testing.T) {
	kubeConfig := "./assets/config"
	uu := []struct {
		flags   *genericclioptions.ConfigFlags
		cluster string
		err     error
	}{
		{&genericclioptions.ConfigFlags{KubeConfig: &kubeConfig}, "blee", nil},
		{&genericclioptions.ConfigFlags{KubeConfig: &kubeConfig}, "bozo", errors.New("invalid context `bozo specified")},
	}

	for _, u := range uu {
		cfg := k8s.NewConfig(u.flags)
		ctx, err := cfg.GetContext(u.cluster)
		if err != nil {
			assert.Equal(t, u.err, err)
		} else {
			assert.NotNil(t, ctx)
			assert.Equal(t, u.cluster, ctx.Cluster)
		}
	}
}

func TestConfigSwitchContext(t *testing.T) {
	cluster, kubeConfig := "duh", "./assets/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := k8s.NewConfig(&flags)
	err := cfg.SwitchContext("blee")
	assert.Nil(t, err)
	ctx, err := cfg.CurrentContextName()
	assert.Nil(t, err)
	assert.Equal(t, "blee", ctx)
}

func TestConfigClusterNameFromContext(t *testing.T) {
	cluster, kubeConfig := "duh", "./assets/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := k8s.NewConfig(&flags)
	cl, err := cfg.ClusterNameFromContext("blee")
	assert.Nil(t, err)
	assert.Equal(t, "blee", cl)
}

func TestConfigAccess(t *testing.T) {
	cluster, kubeConfig := "duh", "./assets/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := k8s.NewConfig(&flags)
	acc, err := cfg.ConfigAccess()
	assert.Nil(t, err)
	assert.True(t, len(acc.GetDefaultFilename()) > 0)
}

func TestConfigContexts(t *testing.T) {
	cluster, kubeConfig := "duh", "./assets/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := k8s.NewConfig(&flags)
	cc, err := cfg.Contexts()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cc))
}

func TestConfigContextNames(t *testing.T) {
	cluster, kubeConfig := "duh", "./assets/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := k8s.NewConfig(&flags)
	cc, err := cfg.ContextNames()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cc))
}

func TestConfigClusterNames(t *testing.T) {
	cluster, kubeConfig := "duh", "./assets/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := k8s.NewConfig(&flags)
	cc, err := cfg.ClusterNames()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cc))
}

func TestConfigDelContext(t *testing.T) {
	cluster, kubeConfig := "duh", "./assets/config.1"
	flags := genericclioptions.ConfigFlags{
		KubeConfig:  &kubeConfig,
		ClusterName: &cluster,
	}

	cfg := k8s.NewConfig(&flags)
	err := cfg.DelContext("fred")
	assert.Nil(t, err)
	cc, err := cfg.ContextNames()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(cc))
}

func TestConfigRestConfig(t *testing.T) {
	kubeConfig := "./assets/config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
	}

	cfg := k8s.NewConfig(&flags)
	rc, err := cfg.RESTConfig()
	assert.Nil(t, err)
	assert.Equal(t, "https://localhost:3000", rc.Host)
}

func TestConfigBadConfig(t *testing.T) {
	kubeConfig := "./assets/bork_config"
	flags := genericclioptions.ConfigFlags{
		KubeConfig: &kubeConfig,
	}

	cfg := k8s.NewConfig(&flags)
	_, err := cfg.RESTConfig()
	assert.NotNil(t, err)
}
