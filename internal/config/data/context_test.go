// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/stretchr/testify/assert"
)

func TestClusterValidate(t *testing.T) {
	c := data.NewContext()
	c.Validate(mock.NewMockConnection(), "ct-1", "cl-1")

	assert.Equal(t, "po", c.View.Active)
	assert.Equal(t, "default", c.Namespace.Active)
	assert.Len(t, c.Namespace.Favorites, 1)
	assert.Equal(t, []string{"default"}, c.Namespace.Favorites)
}

func TestClusterValidateEmpty(t *testing.T) {
	c := data.NewContext()
	c.Validate(mock.NewMockConnection(), "ct-1", "cl-1")

	assert.Equal(t, "po", c.View.Active)
	assert.Equal(t, "default", c.Namespace.Active)
	assert.Len(t, c.Namespace.Favorites, 1)
	assert.Equal(t, []string{"default"}, c.Namespace.Favorites)
}
