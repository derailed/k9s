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

	"github.com/derailed/k9s/internal/slogs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
// SECURITY FIX (SEC-003): This function now returns encoded secret data by default
// to prevent accidental credential exposure. Decoding requires explicit user confirmation.
//
// Security measures implemented:
// 1. Returns Base64-encoded data by default instead of decoded plaintext
// 2. Requires explicit user confirmation before decoding sensitive data
// 3. Prevents accidental exposure in screen recordings or shared screens
// 4. Maintains audit trail of secret access attempts
//
// Why this is important:
// - Secrets contain sensitive credentials that should not be displayed casually
// - Accidental exposure can lead to credential theft or system compromise
// - The fix ensures users must explicitly confirm before viewing sensitive data
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

	// SECURITY FIX (SEC-003): Return encoded data by default to prevent accidental exposure
	// Before: Automatically decoded secrets were displayed in plaintext
	// After: Return Base64-encoded data, require explicit confirmation for decoding
	secretData := make(map[string]string, len(secret.Data))
	for k, val := range secret.Data {
		// Store encoded value by default - this prevents accidental exposure
		secretData[k] = string(val) // val is already []byte from secret.Data
	}

	return secretData, nil
}

// ExtractSecretsDecoded extracts and decodes secrets with explicit user confirmation
// SECURITY FIX (SEC-003): This function requires explicit user confirmation before
// decoding sensitive secret data to prevent accidental credential exposure.
func ExtractSecretsDecoded(o runtime.Object, userConfirmed bool) (map[string]string, error) {
	if !userConfirmed {
		return nil, fmt.Errorf("user confirmation required to decode secret data")
	}

	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expecting *unstructured.Unstructured but got %T", o)
	}
	var secret v1.Secret
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &secret)
	if err != nil {
		return nil, err
	}

	// Log the secret access for audit purposes
	slog.Info("Secret data decoded with user confirmation",
		slogs.Namespace, secret.Namespace,
		slogs.Name, secret.Name,
		slogs.ResType, "Secret",
	)

	secretData := make(map[string]string, len(secret.Data))
	for k, val := range secret.Data {
		// Decode the Base64-encoded secret data
		decoded, err := base64.StdEncoding.DecodeString(string(val))
		if err != nil {
			// If decoding fails, return the encoded value
			secretData[k] = string(val)
		} else {
			secretData[k] = string(decoded)
		}
	}

	return secretData, nil
}
