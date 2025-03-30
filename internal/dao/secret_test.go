// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
)

func TestEncodedSecretDescribe(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.NewGVR("v1/secrets"))

	encodedString := `
Name: bootstrap-token-abcdef
Namespace:    kube-system
Labels:       <none>
Annotations:  <none>

Type:  generic

Data
====
token-secret:  24 bytes`

	expected := "\nName: bootstrap-token-abcdef\n" +
		"Namespace:    kube-system\n" +
		"Labels:       <none>\n" +
		"Annotations:  <none>\n" +
		"\n" +
		"Type:  generic\n" +
		"\n" +
		"Data\n" +
		"====\n" +
		"token-secret: 0123456789abcdef"

	decodedDescription, err := s.Decode(encodedString, "kube-system/bootstrap-token-abcdef")
	assert.NoError(t, err)
	assert.Equal(t, expected, decodedDescription)
}
