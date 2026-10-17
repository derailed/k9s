// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/printers"
)

// Secret represents a secret K8s resource.
type Secret struct {
	Resource
	decodeData bool
}

// Describe describes a secret that can be encoded or decoded.
func (s *Secret) Describe(path string) (string, error) {
	encodedDescription, err := s.Generic.Describe(path)
	if err != nil {
		return "", err
	}
	if s.decodeData {
		return s.Decode(encodedDescription, path)
	}

	return encodedDescription, nil
}

// ToYAML returns a resource yaml.
func (s *Secret) ToYAML(path string, showManaged bool) (string, error) {
	if s.decodeData {
		return s.decodeYAML(path, showManaged)
	}

	return s.Generic.ToYAML(path, showManaged)
}

func (s *Secret) decodeYAML(path string, showManaged bool) (string, error) {
	o, err := s.Get(context.Background(), path)
	if err != nil {
		return "", err
	}
	o = o.DeepCopyObject()
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return "", fmt.Errorf("expecting unstructured but got %T", o)
	}
	if u.Object == nil {
		return "", fmt.Errorf("expecting unstructured object but got nil")
	}
	if !showManaged {
		if meta, ok := u.Object["metadata"].(map[string]any); ok {
			delete(meta, "managedFields")
		}
	}
	if decoded, err := ExtractSecrets(o); err == nil {
		u.Object["data"] = decoded
	}

	var (
		buff bytes.Buffer
		p    printers.YAMLPrinter
	)
	if err := p.PrintObj(o, &buff); err != nil {
		slog.Error("PrintObj failed", slogs.Error, err)
		return "", err
	}

	return buff.String(), nil
}

// SetDecodeData toggles decode mode.
func (s *Secret) SetDecodeData(b bool) {
	s.decodeData = b
}

// Decode removes the encoded part from the secret's description and appends the
// secret's decoded data.
func (s *Secret) Decode(encodedDescription, path string) (string, error) {
	dataEndIndex := strings.Index(encodedDescription, "====")
	if dataEndIndex == -1 {
		return "", fmt.Errorf("unable to find data section in secret description")
	}

	dataEndIndex += 4
	if dataEndIndex >= len(encodedDescription) {
		return "", fmt.Errorf("data section in secret description is invalid")
	}

	// Remove the encoded part from k8s's describe API
	// More details about the reasoning of index: https://github.com/kubernetes/kubectl/blob/v0.29.0/pkg/describe/describe.go#L2542
	body := encodedDescription[0:dataEndIndex]

	o, err := s.Get(context.Background(), path)
	if err != nil {
		return "", err
	}
	data, err := ExtractSecrets(o)
	if err != nil {
		return "", err
	}
	decodedSecrets := make([]string, 0, len(data))
	for _, k := range slices.Sorted(maps.Keys(data)) {
		line := fmt.Sprintf("%s: %s", k, data[k])
		decodedSecrets = append(decodedSecrets, strings.TrimSpace(line))
	}

	return body + "\n" + strings.Join(decodedSecrets, "\n"), nil
}

// ExtractSecrets takes an unstructured object and attempts to convert it into a
// Kubernetes Secret.
// It returns a map where the keys are the secret data keys and the values are
// the corresponding secret data values.
// If the conversion fails, it returns an error.
func ExtractSecrets(o runtime.Object) (map[string]string, error) {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expecting *unstructured.Unstructured but got %T", o)
	}
	var secret v1.Secret
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &secret)
	if err != nil {
		return nil, err
	}
	secretData := make(map[string]string, len(secret.Data))
	for k, val := range secret.Data {
		secretData[k] = string(val)
	}

	return secretData, nil
}

// GetEditableYAML returns the Secret as YAML for the decoded editor.
// A value moves to stringData only when the YAML printer returns the same
// bytes. Every other value stays in data as base64.
func (s *Secret) GetEditableYAML(path string) ([]byte, error) {
	o, err := s.Get(context.Background(), path)
	if err != nil {
		return nil, err
	}
	o = o.DeepCopyObject()
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expecting unstructured but got %T", o)
	}
	if u.Object == nil {
		return nil, fmt.Errorf("expecting unstructured object but got nil")
	}

	if meta, ok := u.Object["metadata"].(map[string]any); ok {
		delete(meta, "managedFields")
	}

	var secret v1.Secret
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &secret); err != nil {
		return nil, fmt.Errorf("failed to convert secret for decoded edit: %w", err)
	}
	// Plaintext goes in stringData only when the YAML printer returns the
	// same bytes. Everything else stays in data as base64.
	stringData := make(map[string]any)
	binaryData := make(map[string]any)
	for k, val := range secret.Data {
		if yamlStringRoundTrips(val) {
			stringData[k] = string(val)
		} else {
			binaryData[k] = base64.StdEncoding.EncodeToString(val)
		}
	}
	if len(binaryData) > 0 {
		u.Object["data"] = binaryData
	} else {
		delete(u.Object, "data")
	}
	if len(stringData) > 0 {
		u.Object["stringData"] = stringData
	} else {
		delete(u.Object, "stringData")
	}

	var (
		buff bytes.Buffer
		p    printers.YAMLPrinter
	)
	if err := p.PrintObj(o, &buff); err != nil {
		return nil, err
	}

	out := make([]byte, 0, len(decodedEditHeader)+buff.Len())
	out = append(out, decodedEditHeader...)
	out = append(out, buff.Bytes()...)

	return out, nil
}

const (
	decodedEditHeader = "# data is base64. stringData is plaintext.\n"
	// coreAPIVersion is the core/v1 apiVersion. A literal "v1" here is one
	// more copy than goconst allows in this package.
	coreAPIVersion = "v1"
)

// yamlStringRoundTrips reports whether val survives the editor's YAML printer.
// Valid UTF-8 is not enough: U+0085 is printed as a space, and some controls
// make the printer fail.
func yamlStringRoundTrips(val []byte) bool {
	if !utf8.Valid(val) {
		return false
	}
	probe := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": coreAPIVersion,
		"kind":       "Secret",
		"stringData": map[string]any{"k": string(val)},
	}}
	var (
		buff bytes.Buffer
		p    printers.YAMLPrinter
	)
	if err := p.PrintObj(probe, &buff); err != nil {
		return false
	}
	var back unstructured.Unstructured
	dec := yaml.NewYAMLOrJSONDecoder(&buff, buff.Len())
	if err := dec.Decode(&back.Object); err != nil || back.Object == nil {
		return false
	}
	sd, ok := back.Object["stringData"].(map[string]any)
	if !ok {
		return false
	}
	got, ok := sd["k"].(string)
	if !ok {
		return false
	}

	return got == string(val)
}

// OpenedResourceVersion returns the resourceVersion in an editor buffer.
// Save pins this value so the buffer cannot clear or replace it.
func OpenedResourceVersion(raw []byte) (string, error) {
	var obj unstructured.Unstructured
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(raw), len(raw))
	if err := dec.Decode(&obj.Object); err != nil || obj.Object == nil {
		return "", fmt.Errorf("opened secret is missing a resourceVersion")
	}
	rv := obj.GetResourceVersion()
	if rv == "" {
		return "", fmt.Errorf("opened secret is missing a resourceVersion")
	}

	return rv, nil
}

// PrepareEditedSecret parses edited YAML, validates data values as
// already-base64, and forces metadata name, namespace, and resourceVersion
// to the opened Secret so the editor cannot retarget it.
func PrepareEditedSecret(path, openedResourceVersion string, editedYAML []byte) (*unstructured.Unstructured, error) {
	var obj unstructured.Unstructured
	bufSize := len(editedYAML)
	if bufSize == 0 {
		bufSize = 64
	}
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(editedYAML), bufSize)
	if err := dec.Decode(&obj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse edited YAML: %w", err)
	}
	if obj.Object == nil {
		return nil, fmt.Errorf("edited document is empty")
	}
	ns, n := client.Namespaced(path)
	if n == "" {
		return nil, fmt.Errorf("missing resource name in path %q", path)
	}
	if obj.GetAPIVersion() != coreAPIVersion || obj.GetKind() != "Secret" {
		return nil, fmt.Errorf("edited document must be apiVersion v1, kind Secret")
	}
	if openedResourceVersion == "" {
		return nil, fmt.Errorf("refusing to update secret %s without a resourceVersion", path)
	}
	if client.IsClusterScoped(ns) {
		ns = client.BlankNamespace
	}
	obj.SetNamespace(ns)
	obj.SetName(n)
	obj.SetResourceVersion(openedResourceVersion)
	if raw, exists := obj.Object["data"]; exists && raw != nil {
		data, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("data must be a map of base64 strings")
		}
		if err := EncodeSecretData(data); err != nil {
			return nil, err
		}
	}
	if err := requireStringData(obj.Object["stringData"]); err != nil {
		return nil, err
	}

	return &obj, nil
}

func requireStringData(raw any) error {
	if raw == nil {
		return nil
	}
	stringData, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("stringData must be a map of strings")
	}
	for k, v := range stringData {
		if _, ok := v.(string); !ok {
			return fmt.Errorf("stringData.%s must be a string", k)
		}
	}

	return nil
}

// UpdateFromEditedYAML parses edited YAML and updates the Secret via
// the K8s API. data values must already be base64; stringData is plaintext.
func (s *Secret) UpdateFromEditedYAML(path, openedResourceVersion string, editedYAML []byte) error {
	obj, err := PrepareEditedSecret(path, openedResourceVersion, editedYAML)
	if err != nil {
		return err
	}

	ns, n := obj.GetNamespace(), obj.GetName()
	auth, err := s.Client().CanI(ns, s.gvr, n, client.UpdateAccess)
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to update secret %s/%s", ns, n)
	}

	dial, err := s.dynClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Client().Config().CallTimeout())
	defer cancel()

	if client.IsClusterScoped(ns) {
		_, err = dial.Update(ctx, obj, metav1.UpdateOptions{})
	} else {
		_, err = dial.Namespace(ns).Update(ctx, obj, metav1.UpdateOptions{})
	}

	return err
}

// EncodeSecretData treats each data value as already-base64 (the same
// contract as kubectl apply). It strips whitespace, rejects values that
// are not valid standard base64, and never encodes again. Put plaintext
// in stringData.
func EncodeSecretData(data map[string]any) error {
	for k, v := range data {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("data.%s must be a base64 string", k)
		}
		compact := strings.Map(func(r rune) rune {
			if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
				return -1
			}
			return r
		}, s)
		if _, err := base64.StdEncoding.DecodeString(compact); err != nil {
			return fmt.Errorf("data.%s is not valid base64; put plaintext in stringData", k)
		}
		data[k] = compact
	}
	return nil
}
