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

func secretUnstructured(name string, data map[string]any) *unstructured.Unstructured {
	obj := map[string]any{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]any{
			"name":            name,
			"namespace":       "default",
			"resourceVersion": "100",
		},
		"type": "Opaque",
	}
	if data != nil {
		obj["data"] = data
	}
	return &unstructured.Unstructured{Object: obj}
}

func secretFactory(name string, data map[string]any) dao.Factory {
	return &testFactory{
		inventory: map[string]map[*client.GVR][]runtime.Object{
			"default": {
				client.SecGVR: {secretUnstructured(name, data)},
			},
		},
	}
}

func TestGetEditableYAML(t *testing.T) {
	b64Binary := base64.StdEncoding.EncodeToString([]byte{0x00, 0x01, 0xFF, 0xFE, 0x80})

	uu := map[string]struct {
		factory     dao.Factory
		path        string
		contains    []string
		notContains []string
	}{
		"basic-decode": {
			factory: makeFactory(),
			path:    "kube-system/bootstrap-token-abcdef",
			contains: []string{
				"token-secret: 0123456789abcdef",
				"kind: Secret",
				"name: bootstrap-token-abcdef",
			},
			notContains: []string{
				"MDEyMzQ1Njc4OWFiY2RlZg==",
				"managedFields",
			},
		},
		"binary-data": {
			factory:  secretFactory("binary-secret", map[string]any{"binary-key": b64Binary}),
			path:     "default/binary-secret",
			contains: []string{b64Binary},
		},
		"empty-data": {
			factory:  secretFactory("empty-secret", nil),
			path:     "default/empty-secret",
			contains: []string{"kind: Secret"},
		},
		"mixed-text-and-binary": {
			factory: secretFactory("mixed-secret", map[string]any{
				"other":  base64.StdEncoding.EncodeToString([]byte("plain")),
				"binary": "AAEB//7wgA==",
			}),
			path: "default/mixed-secret",
			contains: []string{
				"stringData:",
				"other: plain",
				"AAEB//7wgA==",
			},
		},
		"multi-keys": {
			factory: secretFactory("multi-secret", map[string]any{
				"username": base64.StdEncoding.EncodeToString([]byte("admin")),
				"password": base64.StdEncoding.EncodeToString([]byte("s3cr3t")),
			}),
			path: "default/multi-secret",
			contains: []string{
				"username: admin",
				"password: s3cr3t",
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			var s dao.Secret
			s.Init(u.factory, client.SecGVR)
			raw, err := s.GetEditableYAML(u.path)
			require.NoError(t, err)
			y := string(raw)
			for _, c := range u.contains {
				assert.Contains(t, y, c)
			}
			for _, nc := range u.notContains {
				assert.NotContains(t, y, nc)
			}
		})
	}
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

// uozalp on #3982: GetEditableYAML leaves non-UTF-8 as base64. Save must
// not encode that string again (AAEB//7wgA== -> QUFFQi8vN3dnQT09).
func TestEncodeSecretData_binarySiblingUnchanged(t *testing.T) {
	const binaryB64 = "AAEB//7wgA=="
	data := map[string]any{
		"binary": binaryB64,
		"other":  "changed",
	}
	dao.EncodeSecretData(data)
	assert.Equal(t, binaryB64, data["binary"])
	assert.NotEqual(t, "QUFFQi8vN3dnQT09", data["binary"])
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("changed")), data["other"])
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	raw, err := s.GetEditableYAML("kube-system/bootstrap-token-abcdef")
	require.NoError(t, err)

	y := string(raw)
	assert.Contains(t, y, "token-secret: 0123456789abcdef")
	assert.Contains(t, y, "stringData:")
}

func TestUpdateFromEditedYAML_ParseError(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	err := s.UpdateFromEditedYAML([]byte("not: valid: yaml: {{{}"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse edited YAML")
}
