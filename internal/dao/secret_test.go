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
		wantErr     string
	}{
		"basic-decode": {
			factory: makeFactory(),
			path:    "kube-system/bootstrap-token-abcdef",
			contains: []string{
				"token-secret-f: 0123456789abcdef",
				"stringData:",
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
		"convert-error": {
			factory: &testFactory{
				inventory: map[string]map[*client.GVR][]runtime.Object{
					"default": {
						client.SecGVR: {&unstructured.Unstructured{Object: map[string]any{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]any{
								"name":      "bad-secret",
								"namespace": "default",
							},
							"data": "not-a-map",
						}}},
					},
				},
			},
			path:    "default/bad-secret",
			wantErr: "failed to convert secret for decoded edit",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			var s dao.Secret
			s.Init(u.factory, client.SecGVR)
			raw, err := s.GetEditableYAML(u.path)
			if u.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), u.wantErr)
				return
			}
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
	utf8B64 := base64.StdEncoding.EncodeToString([]byte("my-password"))
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr string
	}{
		{name: "utf8-already-base64", input: utf8B64, want: utf8B64},
		{name: "empty-string", input: "", want: ""},
		{name: "binary-already-base64", input: "AAEB//7wgA==", want: "AAEB//7wgA=="},
		{name: "plaintext", input: "my-password", wantErr: "not valid base64"},
		{name: "non-string", input: 42, wantErr: "base64 string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]any{"key": tt.input}
			err := dao.EncodeSecretData(data)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, data["key"])
		})
	}
}

func TestEncodeSecretData_binarySiblingUnchanged(t *testing.T) {
	const binaryB64 = "AAEB//7wgA=="
	data := map[string]any{
		"binary": binaryB64,
		"other":  base64.StdEncoding.EncodeToString([]byte("changed")),
	}
	require.NoError(t, dao.EncodeSecretData(data))
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
	assert.Contains(t, y, "token-secret-f: 0123456789abcdef")
	assert.Contains(t, y, "stringData:")
}

func TestUpdateFromEditedYAML_ParseError(t *testing.T) {
	var s dao.Secret
	s.Init(makeFactory(), client.SecGVR)

	err := s.UpdateFromEditedYAML("default/empty-secret", []byte("not: valid: yaml: {{{}"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse edited YAML")
}

func TestPrepareEditedSecret_pinsOpenedPath(t *testing.T) {
	edited := []byte(`apiVersion: v1
kind: Secret
metadata:
  name: other-secret
  namespace: other-ns
stringData:
  k: v
`)
	obj, err := dao.PrepareEditedSecret("default/empty-secret", edited)
	require.NoError(t, err)
	assert.Equal(t, "empty-secret", obj.GetName())
	assert.Equal(t, "default", obj.GetNamespace())
}

func TestPrepareEditedSecret_missingName(t *testing.T) {
	_, err := dao.PrepareEditedSecret("default/", []byte("kind: Secret"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing resource name")
}

func secretMaps(t *testing.T, obj *unstructured.Unstructured) (data, stringData map[string]any) {
	t.Helper()
	data, _ = obj.Object["data"].(map[string]any)
	stringData, _ = obj.Object["stringData"].(map[string]any)
	if data == nil {
		data = map[string]any{}
	}
	if stringData == nil {
		stringData = map[string]any{}
	}
	return data, stringData
}

// data: values are already base64 (kubectl apply semantics). stringData is plaintext.
func TestPrepareEditedSecret_dataStaysBase64(t *testing.T) {
	plain := "password" + "2"
	plainB64 := base64.StdEncoding.EncodeToString([]byte(plain))
	doubleB64 := base64.StdEncoding.EncodeToString([]byte(plainB64))
	const binaryB64 = "AAEB//7wgA=="

	tests := []struct {
		name       string
		yaml       string
		wantData   map[string]string
		wantString map[string]string
		wantErr    string
		notData    []string
	}{
		{
			name: "uozalp-utf8-base64-in-data",
			yaml: `apiVersion: v1
kind: Secret
metadata:
  name: k9s-secret-encoding-repro
  namespace: default
data:
  rotated-key: cGFzc3dvcmQy
stringData:
  password: password2
`,
			wantData:   map[string]string{"rotated-key": plainB64},
			wantString: map[string]string{"password": "password2"},
			notData:    []string{doubleB64},
		},
		{
			name: "leave-binary-edit-stringdata",
			yaml: `apiVersion: v1
kind: Secret
metadata:
  name: mixed-secret
data:
  binary: AAEB//7wgA==
stringData:
  other: changed
`,
			wantData:   map[string]string{"binary": binaryB64},
			wantString: map[string]string{"other": "changed"},
		},
		{
			name: "replace-binary-with-new-binary",
			yaml: `apiVersion: v1
kind: Secret
metadata:
  name: mixed-secret
data:
  binary: /wAB/g==
`,
			wantData: map[string]string{"binary": "/wAB/g=="},
		},
		{
			name: "stringdata-only",
			yaml: `apiVersion: v1
kind: Secret
metadata:
  name: text-secret
stringData:
  username: admin
  password: s3cr3t
`,
			wantString: map[string]string{"username": "admin", "password": "s3cr3t"},
		},
		{
			name: "empty-data-value",
			yaml: `apiVersion: v1
kind: Secret
metadata:
  name: empty-val
data:
  blank: ""
`,
			wantData: map[string]string{"blank": ""},
		},
		{
			name:     "wrapped-base64",
			yaml:     "apiVersion: v1\nkind: Secret\nmetadata:\n  name: wrap\ndata:\n  k: cGFz\n    c3dvcmQy\n",
			wantData: map[string]string{"k": plainB64},
		},
		{
			name: "token-looks-like-base64-in-data",
			yaml: `apiVersion: v1
kind: Secret
metadata:
  name: token
data:
  token-secret-f: MDEyMzQ1Njc4OWFiY2RlZg==
`,
			wantData: map[string]string{"token-secret-f": "MDEyMzQ1Njc4OWFiY2RlZg=="},
		},
		{
			name: "token-plaintext-in-stringdata",
			yaml: `apiVersion: v1
kind: Secret
metadata:
  name: token
stringData:
  token-secret-f: 0123456789abcdef
`,
			wantString: map[string]string{"token-secret-f": "0123456789abcdef"},
		},
		{
			name: "same-key-in-both",
			yaml: `apiVersion: v1
kind: Secret
metadata:
  name: both
data:
  k: cGFzc3dvcmQy
stringData:
  k: password2
`,
			wantData:   map[string]string{"k": plainB64},
			wantString: map[string]string{"k": "password2"},
		},
		{
			name: "add-new-data-key",
			yaml: `apiVersion: v1
kind: Secret
metadata:
  name: add
data:
  rotated-key: AAEB//7wgA==
  extra: cGFzc3dvcmQy
stringData:
  password: password2
`,
			wantData:   map[string]string{"rotated-key": binaryB64, "extra": plainB64},
			wantString: map[string]string{"password": "password2"},
		},
		{
			name: "delete-data-key",
			yaml: `apiVersion: v1
kind: Secret
metadata:
  name: del
stringData:
  password: password2
`,
			wantString: map[string]string{"password": "password2"},
		},
		{
			name:    "plaintext-in-data",
			yaml:    "apiVersion: v1\nkind: Secret\nmetadata:\n  name: bad\ndata:\n  k: password2\n",
			wantErr: "not valid base64",
		},
		{
			name:    "invalid-base64-in-data",
			yaml:    "apiVersion: v1\nkind: Secret\nmetadata:\n  name: bad\ndata:\n  k: not!!!base64\n",
			wantErr: "not valid base64",
		},
		{
			name:    "non-string-data-value",
			yaml:    "apiVersion: v1\nkind: Secret\nmetadata:\n  name: bad\ndata:\n  k: 12345\n",
			wantErr: "base64 string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := dao.PrepareEditedSecret("default/"+tt.name, []byte(tt.yaml))
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			data, stringData := secretMaps(t, obj)
			for k, v := range tt.wantData {
				assert.Equal(t, v, data[k], "data.%s", k)
			}
			for k, v := range tt.wantString {
				assert.Equal(t, v, stringData[k], "stringData.%s", k)
			}
			for _, bad := range tt.notData {
				for _, v := range data {
					assert.NotEqual(t, bad, v)
				}
			}
			if len(tt.wantData) == 0 {
				_, ok := obj.Object["data"]
				assert.False(t, ok, "data map should be absent")
			}
		})
	}
}

func TestPrepareEditedSecret_roundtripUneditedMixed(t *testing.T) {
	var s dao.Secret
	s.Init(secretFactory("mixed-secret", map[string]any{
		"other":  base64.StdEncoding.EncodeToString([]byte("plain")),
		"binary": "AAEB//7wgA==",
	}), client.SecGVR)

	raw, err := s.GetEditableYAML("default/mixed-secret")
	require.NoError(t, err)

	obj, err := dao.PrepareEditedSecret("default/mixed-secret", raw)
	require.NoError(t, err)
	data, stringData := secretMaps(t, obj)
	assert.Equal(t, "AAEB//7wgA==", data["binary"])
	assert.NotContains(t, data, "other")
	assert.Equal(t, "plain", stringData["other"])
}
