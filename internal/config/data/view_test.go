// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/stretchr/testify/assert"
)

func TestViewValidate(t *testing.T) {
	v := data.NewView()

	v.Validate()
	assert.Equal(t, "po", v.Active)

	v.Active = "fred"
	v.Validate()
	assert.Equal(t, "fred", v.Active)
}

func TestViewValidateBlank(t *testing.T) {
	var v data.View
	v.Validate()
	assert.Equal(t, "po", v.Active)
}
