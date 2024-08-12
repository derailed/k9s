// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
)

func TestEncodedSecretDescribe(t *testing.T) {
	s := dao.Secret{}
	s.Init(makeFactory(), client.NewGVR("v1/secrets"))

	encodedString :=
		`
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
		"token-secret:\t0123456789abcdef"

	decodedDescription, _ := s.Decode(encodedString, "kube-system/bootstrap-token-abcdef")
	assert.Equal(t, expected, decodedDescription)
}
