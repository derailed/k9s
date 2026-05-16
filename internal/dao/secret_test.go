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
			factory: &testFactory{
				inventory: map[string]map[*client.GVR][]runtime.Object{
					"default": {
						client.SecGVR: {&unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "v1",
								"kind":       "Secret",
								"metadata": map[string]any{
									"name":            "binary-secret",
									"namespace":       "default",
									"resourceVersion": "100",
								},
								"data": map[string]any{
									"binary-key": b64Binary,
								},
								"type": "Opaque",
							},
						}},
					},
				},
			},
			path:     "default/binary-secret",
			contains: []string{b64Binary},
		},
		"empty-data": {
			factory: &testFactory{
				inventory: map[string]map[*client.GVR][]runtime.Object{
					"default": {
						client.SecGVR: {&unstructured.Unstructured{
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
						}},
					},
				},
			},
			path:     "default/empty-secret",
			contains: []string{"kind: Secret"},
		},
		"multi-keys": {
			factory: &testFactory{
				inventory: map[string]map[*client.GVR][]runtime.Object{
					"default": {
						client.SecGVR: {&unstructured.Unstructured{
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
						}},
					},
				},
			},
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

func TestUpdateFromEditedYAML_ParseError(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	err := s.UpdateFromEditedYAML([]byte("not: valid: yaml: {{{}"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse edited YAML")
}
