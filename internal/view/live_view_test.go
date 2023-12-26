// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/stretchr/testify/assert"
)

func TestLiveViewSetText(t *testing.T) {
	s := `
apiVersion: v1
  data:
    the secret name you want to quote to use tls.","title":"secretName","type":"string"}},"required":["http","class","classInSpec"],"type":"object"}
`

	v := NewLiveView(NewApp(mock.NewMockConfig()), "fred", nil)
	assert.NoError(t, v.Init(context.Background()))
	v.text.SetText(colorizeYAML(config.Yaml{}, s))

	assert.Equal(t, s, sanitizeEsc(v.text.GetText(true)))
}
