// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestPVCNew(t *testing.T) {
	v := view.NewPersistentVolumeClaim(client.NewGVR("v1/persistentvolumeclaims"))

	assert.Nil(t, v.Init(makeCtx()))
	assert.Equal(t, "PersistentVolumeClaims", v.Name())
	assert.Equal(t, 11, len(v.Hints()))
}
