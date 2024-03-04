// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Secret represents a secret K8s resource.
type Secret struct {
	Resource
	decode bool
}

// Describe describes a secret that can be encoded or decoded.
func (s *Secret) Describe(path string) (string, error) {
	encodedDescription, err := s.Generic.Describe(path)

	if err != nil {
		return "", err
	}

	if !s.decode {
		return encodedDescription, nil
	}

	return s.Decode(encodedDescription, path)
}

// SetDecode sets the decode flag.
func (s *Secret) SetDecode(flag bool) {
	s.decode = flag
}

// Decode removes the encoded part from the secret's description and appends the
// secret's decoded data.
func (s *Secret) Decode(encodedDescription, path string) (string, error) {
	o, err := s.getFactory().Get(s.GVR(), path, true, labels.Everything())

	if err != nil {
		return "", err
	}

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

	d, err := ExtractSecrets(o.(*unstructured.Unstructured))

	if err != nil {
		return "", err
	}

	decodedSecrets := []string{}

	for k, v := range d {
		decodedSecrets = append(decodedSecrets, "\n", k, ":\t", v)
	}

	return body + strings.Join(decodedSecrets, ""), nil
}

// ExtractSecrets takes an unstructured object and attempts to convert it into a
// Kubernetes Secret.
// It returns a map where the keys are the secret data keys and the values are
// the corresponding secret data values.
// If the conversion fails, it returns an error.
func ExtractSecrets(o *unstructured.Unstructured) (map[string]string, error) {
	var secret v1.Secret
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, &secret)

	if err != nil {
		return nil, err
	}

	secretData := make(map[string]string, len(secret.Data))

	for k, val := range secret.Data {
		secretData[k] = string(val)
	}

	return secretData, nil
}
