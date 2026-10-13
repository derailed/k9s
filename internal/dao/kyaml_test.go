// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLToKYAML(t *testing.T) {
	raw := `apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  labels:
    app: demo
  annotations:
    foo/bar: "1"
  creationTimestamp: null
spec:
  containers:
  - name: nginx
    image: nginx:1.20
    ports:
    - containerPort: 80
    env: []
  replicas: 3
  enabled: true
`
	e := `---
{
  apiVersion: "v1",
  kind: "Pod",
  metadata: {
    name: "my-pod",
    labels: {
      app: "demo"
    },
    annotations: {
      foo/bar: "1"
    },
    creationTimestamp: null
  },
  spec: {
    containers: [
      {
        name: "nginx",
        image: "nginx:1.20",
        ports: [
          {
            containerPort: 80
          }
        ],
        env: []
      }
    ],
    replicas: 3,
    enabled: true
  }
}
`

	kyaml, err := YAMLToKYAML(raw)
	require.NoError(t, err)
	//nolint:testifylint // exact KYAML formatting (flow style, indentation, quoting) is the behavior under test.
	assert.Equal(t, e, kyaml)
}

func TestYAMLToKYAMLMultiDoc(t *testing.T) {
	raw := `apiVersion: v1
kind: Pod
metadata:
  name: p1
---
apiVersion: v1
kind: Pod
metadata:
  name: p2
`
	e := `---
{
  apiVersion: "v1",
  kind: "Pod",
  metadata: {
    name: "p1"
  }
}
---
{
  apiVersion: "v1",
  kind: "Pod",
  metadata: {
    name: "p2"
  }
}
`

	kyaml, err := YAMLToKYAML(raw)
	require.NoError(t, err)
	//nolint:testifylint // exact KYAML formatting (flow style, indentation, quoting) is the behavior under test.
	assert.Equal(t, e, kyaml)
}

func TestYAMLToKYAMLQuotesUnsafeKeys(t *testing.T) {
	raw := `"weird key": "1"
plain-key/ok.2: v
`
	e := `---
{
  "weird key": "1",
  plain-key/ok.2: "v"
}
`

	kyaml, err := YAMLToKYAML(raw)
	require.NoError(t, err)
	//nolint:testifylint // exact KYAML formatting (flow style, indentation, quoting) is the behavior under test.
	assert.Equal(t, e, kyaml)
}

func TestYAMLToKYAMLEmpty(t *testing.T) {
	kyaml, err := YAMLToKYAML("")
	require.NoError(t, err)
	assert.Empty(t, kyaml)
}
