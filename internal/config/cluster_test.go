// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestClusterValidate(t *testing.T) {
	c := config.NewCluster()
	c.Validate(newMockConnection(), newMockKubeSettings(&genericclioptions.ConfigFlags{}))

	assert.Equal(t, "po", c.View.Active)
	assert.Equal(t, "default", c.Namespace.Active)
	assert.Equal(t, 1, len(c.Namespace.Favorites))
	assert.Equal(t, []string{"default"}, c.Namespace.Favorites)
}

func TestClusterValidateEmpty(t *testing.T) {
	c := config.NewCluster()
	c.Validate(newMockConnection(), newMockKubeSettings(&genericclioptions.ConfigFlags{}))

	assert.Equal(t, "po", c.View.Active)
	assert.Equal(t, "default", c.Namespace.Active)
	assert.Equal(t, 1, len(c.Namespace.Favorites))
	assert.Equal(t, []string{"default"}, c.Namespace.Favorites)
}

func namespaces() []v1.Namespace {
	return []v1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fred",
			},
		},
	}
}
