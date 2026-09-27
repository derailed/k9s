// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	restclient "k8s.io/client-go/rest"
)

func TestSecretListRequestsPartialObjectMetadata(t *testing.T) {
	requests := make(chan *http.Request, 1)
	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		requests <- r.Clone(context.Background())
		body := `{
  "apiVersion":"v1",
  "kind":"SecretList",
  "items":[{
    "apiVersion":"v1",
    "kind":"Secret",
    "metadata":{"name":"database-credentials","namespace":"default"},
    "type":"Opaque",
    "data":{"password":"c2VjcmV0"}
  }]
}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})

	connection := &metadataConn{conn: makeConn(), config: &restclient.Config{Host: "https://kubernetes.test", Transport: transport}}
	factory := metadataFactory{podFactory: podFactory{}, connection: connection}
	var secret dao.Secret
	secret.Init(factory, client.SecGVR)
	ctx := context.WithValue(context.Background(), internal.KeyLabels, labels.SelectorFromSet(labels.Set{"app": "db"}))
	ctx = context.WithValue(ctx, internal.KeyFields, "metadata.name=database-credentials")

	objects, err := secret.List(ctx, "default")
	require.NoError(t, err)
	require.Len(t, objects, 1)
	metadata, ok := objects[0].(*metav1.PartialObjectMetadata)
	require.True(t, ok)
	assert.Equal(t, "database-credentials", metadata.Name)

	request := <-requests
	assert.Equal(t, "/api/v1/namespaces/default/secrets", request.URL.Path)
	assert.Contains(t, request.Header.Get("Accept"), "as=PartialObjectMetadataList")
	assert.Equal(t, "app=db", request.URL.Query().Get("labelSelector"))
	assert.Equal(t, "metadata.name=database-credentials", request.URL.Query().Get("fieldSelector"))
	assert.False(t, strings.Contains(fmt.Sprintf("%#v", metadata), "c2VjcmV0"))
}

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

type metadataConn struct {
	*conn
	config *restclient.Config
}

func (c *metadataConn) RestConfig() (*restclient.Config, error) { return c.config, nil }

type metadataFactory struct {
	podFactory
	connection client.Connection
}

func (f metadataFactory) Client() client.Connection { return f.connection }

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
