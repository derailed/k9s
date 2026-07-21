// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/stretchr/testify/assert"
)

func TestGatewayHeader(t *testing.T) {
	gw := Gateway{}
	assert.NotNil(t, gw)
	assert.NotNil(t, gw.Base)

	header := gw.Header("")
	assert.Equal(t, 8, len(header))
}

func TestGatewayRender(t *testing.T) {
	gw := Gateway{}
	assert.NotNil(t, gw)

	row := &model1.Row{}
	assert.NotNil(t, row)
}