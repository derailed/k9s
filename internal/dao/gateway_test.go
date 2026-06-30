// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGateway(t *testing.T) {
	gw := &Gateway{}
	assert.NotNil(t, gw)
	assert.NotNil(t, &gw.Resource)
}
