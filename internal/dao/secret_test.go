// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"encoding/base64"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

func TestGetEditableYAML(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	raw, err := s.GetEditableYAML("kube-system/bootstrap-token-abcdef")
	require.NoError(t, err)

	yaml := string(raw)
	assert.Contains(t, yaml, "token-secret: 0123456789abcdef", "data value should be decoded from base64")
	assert.NotContains(t, yaml, "MDEyMzQ1Njc4OWFiY2RlZg==", "base64-encoded value should not appear")
	assert.Contains(t, yaml, "kind: Secret")
	assert.Contains(t, yaml, "name: bootstrap-token-abcdef")
	assert.NotContains(t, yaml, "managedFields")
}

func TestGetEditableYAML_BinaryData(t *testing.T) {
	binaryVal := []byte{0x00, 0x01, 0xFF, 0xFE, 0x80}
	b64 := base64.StdEncoding.EncodeToString(binaryVal)

	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]any{
				"name":            "binary-secret",
				"namespace":       "default",
				"resourceVersion": "100",
			},
			"data": map[string]any{
				"binary-key": b64,
			},
			"type": "Opaque",
		},
	}

	factory := &testFactory{
		inventory: map[string]map[*client.GVR][]runtime.Object{
			"default": {
				client.SecGVR: {obj},
			},
		},
	}

	var s dao.Secret
	s.Init(factory, client.SecGVR)

	raw, err := s.GetEditableYAML("default/binary-secret")
	require.NoError(t, err)

	yaml := string(raw)
	// Binary data should remain base64-encoded since it's not valid UTF-8
	assert.Contains(t, yaml, b64, "binary data should stay base64-encoded")
}

func TestEncodeSecretData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "my-password",
			expected: base64.StdEncoding.EncodeToString([]byte("my-password")),
		},
		{
			name:     "empty string",
			input:    "",
			expected: base64.StdEncoding.EncodeToString([]byte("")),
		},
		{
			name:     "special characters",
			input:    "p@$$w0rd!#%",
			expected: base64.StdEncoding.EncodeToString([]byte("p@$$w0rd!#%")),
		},
		{
			name:     "multiline",
			input:    "line1\nline2\nline3",
			expected: base64.StdEncoding.EncodeToString([]byte("line1\nline2\nline3")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]any{"key": tt.input}
			dao.EncodeSecretData(data)
			assert.Equal(t, tt.expected, data["key"])
		})
	}
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	raw, err := s.GetEditableYAML("kube-system/bootstrap-token-abcdef")
	require.NoError(t, err)

	yaml := string(raw)
	assert.Contains(t, yaml, "token-secret: 0123456789abcdef")

	// Simulate re-encoding: parse the data section and encode it
	data := map[string]any{"token-secret": "0123456789abcdef"}
	dao.EncodeSecretData(data)
	assert.Equal(t, "MDEyMzQ1Njc4OWFiY2RlZg==", data["token-secret"],
		"roundtrip should produce the original base64 value")
}

func TestGetEditableYAML_EmptyData(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]any{
				"name":            "empty-secret",
				"namespace":       "default",
				"resourceVersion": "100",
			},
			"type": "Opaque",
		},
	}

	factory := &testFactory{
		inventory: map[string]map[*client.GVR][]runtime.Object{
			"default": {
				client.SecGVR: {obj},
			},
		},
	}

	var s dao.Secret
	s.Init(factory, client.SecGVR)

	raw, err := s.GetEditableYAML("default/empty-secret")
	require.NoError(t, err)
	assert.Contains(t, string(raw), "kind: Secret")
}

func TestGetEditableYAML_MultipleKeys(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]any{
				"name":            "multi-secret",
				"namespace":       "default",
				"resourceVersion": "100",
			},
			"data": map[string]any{
				"username": base64.StdEncoding.EncodeToString([]byte("admin")),
				"password": base64.StdEncoding.EncodeToString([]byte("s3cr3t")),
			},
			"type": "Opaque",
		},
	}

	factory := &testFactory{
		inventory: map[string]map[*client.GVR][]runtime.Object{
			"default": {
				client.SecGVR: {obj},
			},
		},
	}

	var s dao.Secret
	s.Init(factory, client.SecGVR)

	raw, err := s.GetEditableYAML("default/multi-secret")
	require.NoError(t, err)

	yaml := string(raw)
	assert.Contains(t, yaml, "username: admin")
	assert.Contains(t, yaml, "password: s3cr3t")
}

func TestUpdateFromEditedYAML_ParseError(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	err := s.UpdateFromEditedYAML([]byte("not: valid: yaml: {{{}"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse edited YAML")
}
