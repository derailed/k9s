// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeProgressBar(t *testing.T) {
	assert.Equal(t, "[#####-----]", sanitizeProgressBar(10, 5, 10))
	assert.Equal(t, "[##########]", sanitizeProgressBar(10, 12, 10))
}

func TestSanitizeProgressText(t *testing.T) {
	assert.Contains(t, sanitizeProgressText(-1, 0, ""), "Scanning pods")
	assert.Contains(t, sanitizeProgressText(0, 0, ""), "Nothing to sanitize")

	msg := sanitizeProgressText(2, 1, "default/pod-0")
	assert.Contains(t, msg, "1/2")
	assert.Contains(t, msg, "Current: default/pod-0")
	assert.Contains(t, msg, "[#")
}