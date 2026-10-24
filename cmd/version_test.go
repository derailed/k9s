// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestClientAppVersionFromCmd(t *testing.T) {
	assert.Equal(t, version, client.AppVersion)
}
