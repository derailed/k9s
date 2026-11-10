// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestClientVersion(t *testing.T) {
	oldClientVersion := version
	expectedVersion := "cmd-pkg-test-version"
	setClientVersion(expectedVersion)

	// validate that setting version via cmd package helper propagates to client.Version
	assert.Equal(t, "cmd-pkg-test-version", client.Version)

	//cleanup
	setClientVersion(oldClientVersion)
}
