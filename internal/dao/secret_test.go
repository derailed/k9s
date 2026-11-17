// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodedSecretDescribe(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	encodedString := `
Name: bootstrap-token-abcdef
Namespace:    kube-system
Labels:       <none>
Annotations:  <none>

Type:  generic

Data
====
token-secret-a:  16 bytes
token-secret-b:  16 bytes
token-secret-c:  16 bytes
token-secret-d:  16 bytes
token-secret-e:  16 bytes
token-secret-f:  16 bytes
token-secret-g:  16 bytes
token-secret-h:  16 bytes`

	expected := "\nName: bootstrap-token-abcdef\n" +
		"Namespace:    kube-system\n" +
		"Labels:       <none>\n" +
		"Annotations:  <none>\n" +
		"\n" +
		"Type:  generic\n" +
		"\n" +
		"Data\n" +
		"====\n" +
		"token-secret-a: 0123456789abcdea\n" +
		"token-secret-b: 0123456789abcdeb\n" +
		"token-secret-c: 0123456789abcdec\n" +
		"token-secret-d: 0123456789abcded\n" +
		"token-secret-e: 0123456789abcdee\n" +
		"token-secret-f: 0123456789abcdef\n" +
		"token-secret-g: 0123456789abcdeg\n" +
		"token-secret-h: 0123456789abcdeh"

	decodedDescription, err := s.Decode(encodedString, "kube-system/bootstrap-token-abcdef")
	require.NoError(t, err)
	assert.Equal(t, expected, decodedDescription)
}
