// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestViewValidate(t *testing.T) {
	v := config.NewView()

	v.Validate()
	assert.Equal(t, "po", v.Active)

	v.Active = "fred"
	v.Validate()
	assert.Equal(t, "fred", v.Active)
}

func TestViewValidateBlank(t *testing.T) {
	var v config.View
	v.Validate()
	assert.Equal(t, "po", v.Active)
}
