// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/derailed/k9s/internal/slogs"
	yamlv3 "gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	sigsyaml "sigs.k8s.io/yaml"
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

// secretKeyOrder surfaces the secret type (for context) followed by the editable
// fields first. sigs.k8s.io/yaml sorts alphabetically (via JSON), which would
// push stringData off-screen and make editing non-obvious.
var secretKeyOrder = []string{"type", "stringData", "data"}

// EditYAML returns the secret YAML decoded for editing. Text values are moved
// to stringData and the output is reordered so editable fields appear first.
func (s *Secret) EditYAML(path string) (string, error) {
	o, err := s.Get(context.Background(), path)
	if err != nil {
		return "", err
	}
	if o == nil {
		return "", fmt.Errorf("secret not found: %s", path)
	}

	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return "", fmt.Errorf("expecting unstructured but got %T", o)
	}

	secret, err := toSecret(o)
	if err != nil {
		return "", fmt.Errorf("failed to convert to secret: %w", err)
	}

	// TypeMeta is not populated by DefaultUnstructuredConverter.
	secret.APIVersion = u.GetAPIVersion()
	secret.Kind = u.GetKind()

	// Separate text-safe values (stringData) from binary values (data).
	// sigs.k8s.io/yaml marshals []byte via JSON, which base64-encodes them.
	stringData := make(map[string]string, len(secret.Data))
	binaryData := make(map[string][]byte, len(secret.Data))
	for k, val := range secret.Data {
		if utf8.Valid(val) {
			stringData[k] = string(val)
		} else {
			binaryData[k] = val
		}
	}
	secret.Data = binaryData
	secret.StringData = stringData

	out, err := sigsyaml.Marshal(secret)
	if err != nil {
		return "", err
	}

	return reorderSecretYAML(string(out))
}

// reorderSecretYAML reorders top-level keys so type (for context) and the
// editable fields (stringData, data) appear first, visible without scrolling.
func reorderSecretYAML(src string) (string, error) {
	var doc yamlv3.Node
	if err := yamlv3.Unmarshal([]byte(src), &doc); err != nil {
		return src, err
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yamlv3.MappingNode {
		return src, nil
	}
	root := doc.Content[0]

	pairs := make(map[string][2]*yamlv3.Node, len(root.Content)/2)
	for i := 0; i+1 < len(root.Content); i += 2 {
		pairs[root.Content[i].Value] = [2]*yamlv3.Node{root.Content[i], root.Content[i+1]}
	}

	newContent := make([]*yamlv3.Node, 0, len(root.Content))
	seen := make(map[string]bool, len(secretKeyOrder))
	for _, k := range secretKeyOrder {
		if p, ok := pairs[k]; ok {
			newContent = append(newContent, p[0], p[1])
			seen[k] = true
		}
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if k := root.Content[i].Value; !seen[k] {
			newContent = append(newContent, root.Content[i], root.Content[i+1])
		}
	}
	root.Content = newContent

	out, err := yamlv3.Marshal(root)
	if err != nil {
		return src, err
	}
	return string(out), nil
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
	for k, v := range data {
		line := fmt.Sprintf("%s: %s", k, v)
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
	secret, err := toSecret(o)
	if err != nil {
		return nil, err
	}
	secretData := make(map[string]string, len(secret.Data))
	for k, val := range secret.Data {
		secretData[k] = string(val)
	}

	return secretData, nil
}

func toSecret(o runtime.Object) (*v1.Secret, error) {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expecting *unstructured.Unstructured but got %T", o)
	}
	var secret v1.Secret
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &secret)
	if err != nil {
		return nil, err
	}

	return &secret, nil
}
