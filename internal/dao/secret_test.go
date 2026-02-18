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
	require.NoError(t, err)
	assert.Equal(t, expected, decodedDescription)
}

func TestEditYAML_AllTextData(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	yaml, err := s.EditYAML("kube-system/bootstrap-token-abcdef")
	require.NoError(t, err)

	// Text data should be in stringData
	assert.Contains(t, yaml, "stringData:")
	assert.Contains(t, yaml, "token-secret: 0123456789abcdef")
	// No data field since all values are text
	assert.NotContains(t, yaml, "\ndata:")
}

func TestEditYAML_MixedData(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	yaml, err := s.EditYAML("default/mixed-secret")
	require.NoError(t, err)

	// Text fields go to stringData
	assert.Contains(t, yaml, "stringData:")
	assert.Contains(t, yaml, "username: admin")
	assert.Contains(t, yaml, "password: s3cr3t")

	// Binary field stays base64-encoded in data
	assert.Contains(t, yaml, "data:")
	assert.Contains(t, yaml, "key: gA4CAwQ=")
}

func TestEditYAML_AllBinaryData(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	yaml, err := s.EditYAML("default/binary-secret")
	require.NoError(t, err)

	// Binary data stays in data field
	assert.Contains(t, yaml, "data:")
	assert.Contains(t, yaml, "binary-key: gA4CAwQ=")
	// No stringData since nothing is text
	assert.NotContains(t, yaml, "stringData:")
}

func TestEditYAML_EmptyData(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	yaml, err := s.EditYAML("default/empty-secret")
	require.NoError(t, err)

	// No stringData or data sections with values
	assert.NotContains(t, yaml, "stringData:")
	assert.Contains(t, yaml, "kind: Secret")
}

func TestExtractSecrets(t *testing.T) {
	o := load("secret_mixed")

	data, err := dao.ExtractSecrets(o)
	require.NoError(t, err)

	assert.Len(t, data, 3)
	assert.Equal(t, "admin", data["username"])
	assert.Equal(t, "s3cr3t", data["password"])
	// Binary data is still returned as raw string (not base64)
	assert.Contains(t, data, "key")
}

func TestEditYAML_SpecialCharacters(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	yaml, err := s.EditYAML("default/special-secret")
	require.NoError(t, err)

	// All valid UTF-8 should be in stringData
	assert.Contains(t, yaml, "stringData:")
	assert.Contains(t, yaml, "multiline:")
	assert.Contains(t, yaml, "line1")
	assert.Contains(t, yaml, "unicode:")
	// YAML printer escapes emoji as Unicode escape sequence
	assert.Regexp(t, `(😊|\\U0001F60A)`, yaml)
	assert.Contains(t, yaml, "whitespace:")
	// No data field since all are valid UTF-8
	assert.NotContains(t, yaml, "\ndata:")
}

func TestEditYAML_NonExistentSecret(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	_, err := s.EditYAML("default/does-not-exist")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
