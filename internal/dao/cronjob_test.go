// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToggledSuspend(t *testing.T) {
	yes, no := true, false

	assert.True(t, toggledSuspend(nil))
	assert.True(t, toggledSuspend(&no))
	assert.False(t, toggledSuspend(&yes))
}
