// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newStyle(t *testing.T) {
	s := newSkin()

	assert.Equal(t, Color("black"), s.K9s.Body.BgColor)
	assert.Equal(t, Color("cadetblue"), s.K9s.Body.FgColor)
	assert.Equal(t, Color("lightskyblue"), s.K9s.Frame.Status.NewColor)
}
