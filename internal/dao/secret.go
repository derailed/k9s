// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
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

// GetEditableYAML returns the full Secret as YAML with data values decoded from
// base64 to plaintext. Values that are not valid UTF-8 (binary) are kept as
// base64 to prevent data corruption.
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
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &secret); err == nil {
		decoded := make(map[string]any, len(secret.Data))
		for k, val := range secret.Data {
			if utf8.Valid(val) {
				decoded[k] = string(val)
			} else {
				decoded[k] = base64.StdEncoding.EncodeToString(val)
			}
		}
		if len(decoded) > 0 {
			u.Object["data"] = decoded
		}
	}

	var (
		buff bytes.Buffer
		p    printers.YAMLPrinter
	)
	if err := p.PrintObj(o, &buff); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// UpdateFromEditedYAML parses edited YAML (with decoded plaintext data values),
// re-encodes data values to base64, and updates the Secret via the K8s API.
func (s *Secret) UpdateFromEditedYAML(editedYAML []byte) error {
	var obj unstructured.Unstructured
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(editedYAML), len(editedYAML))
	if err := dec.Decode(&obj.Object); err != nil {
		return fmt.Errorf("failed to parse edited YAML: %w", err)
	}

	if data, ok := obj.Object["data"].(map[string]any); ok {
		EncodeSecretData(data)
	}

	ns, n := client.Namespaced(obj.GetNamespace() + "/" + obj.GetName())
	auth, err := s.Client().CanI(ns, s.gvr, n, []string{client.UpdateVerb})
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
		_, err = dial.Update(ctx, &obj, metav1.UpdateOptions{})
	} else {
		_, err = dial.Namespace(ns).Update(ctx, &obj, metav1.UpdateOptions{})
	}

	return err
}

// EncodeSecretData base64-encodes all string values in a secret data map
// in-place. This converts decoded plaintext values back to the base64 format
// expected by the Kubernetes API.
func EncodeSecretData(data map[string]any) {
	for k, v := range data {
		s, ok := v.(string)
		if !ok {
			continue
		}
		data[k] = base64.StdEncoding.EncodeToString([]byte(s))
	}
}
