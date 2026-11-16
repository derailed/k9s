// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package watch

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

type dynConn struct {
	client.Connection
	dial dynamic.Interface
}

func (c dynConn) DynDial() (dynamic.Interface, error)                    { return c.dial, nil }
func (dynConn) CanI(string, *client.GVR, string, []string) (bool, error) { return true, nil }

func TestFactoryGetColdCache(t *testing.T) {
	rs := &unstructured.Unstructured{}
	rs.SetAPIVersion("apps/v1")
	rs.SetKind("ReplicaSet")
	rs.SetNamespace("ns1")
	rs.SetName("rs1")

	dial := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{client.RsGVR.GVR(): "ReplicaSetList"},
		rs,
	)
	// Informer LIST never succeeds, so the cache stays cold.
	dial.PrependReactor("list", "replicasets", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("slow list")
	})

	f := NewFactory(dynConn{dial: dial})
	f.Start("ns1")
	defer f.Terminate()

	o, err := f.Get(client.RsGVR, "ns1/rs1", true, nil)
	require.NoError(t, err)
	assert.Equal(t, "rs1", o.(*unstructured.Unstructured).GetName())
}
