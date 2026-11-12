// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestClientVersion(t *testing.T) {
	oldClientVersion := client.Version
	defer func() {
		setClientVersion(oldClientVersion)
	}()

	expectedVersion := "cmd-pkg-test-version"

	// validate that setting version via cmd package helper propagates to client.Version
	setClientVersion(expectedVersion)
	assert.Equal(t, expectedVersion, client.Version)
}
