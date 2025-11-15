// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s
package xray_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/xray"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNodeRender(t *testing.T) {
	// Prepare factory rows: one pod scheduled on node "minikube"
	f := makeFactory()
	f.rows = map[*client.GVR][]runtime.Object{
		client.PodGVR: {load(t, "po")},
	}

	var re xray.Node
	o := load(t, "node")

	root := xray.NewTreeNode(client.NodeGVR, "nodes")
	ctx := context.WithValue(context.Background(), xray.KeyParent, root)
	ctx = context.WithValue(ctx, internal.KeyFactory, f)

	require.NoError(t, re.Render(ctx, "", o))
	// One node entry under root
	assert.Equal(t, 1, root.CountChildren())
	// Node has one namespace grouping (from pod renderer)
	assert.Equal(t, 1, root.Children[0].CountChildren())
	// That namespace has one pod
	assert.Equal(t, 1, root.Children[0].Children[0].CountChildren())
}

func TestNodeRender_PodPVCExpandsToPV(t *testing.T) {
	// Prepare factory rows: one pod on node plus a bound PVC and its PV.
	f := makeFactory()
	f.rows = map[*client.GVR][]runtime.Object{
		client.PodGVR: {load(t, "po")},
		client.PvcGVR: {
			&v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web",
					Namespace: "default",
				},
				Spec: v1.PersistentVolumeClaimSpec{
					VolumeName: "pv-web",
				},
			},
		},
		client.PvGVR: {
			&v1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pv-web",
				},
			},
		},
	}

	var re xray.Node
	o := load(t, "node")

	root := xray.NewTreeNode(client.NodeGVR, "nodes")
	ctx := context.WithValue(context.Background(), xray.KeyParent, root)
	ctx = context.WithValue(ctx, internal.KeyFactory, f)

	require.NoError(t, re.Render(ctx, "", o))
	require.Equal(t, 1, root.CountChildren())
	nodeEntry := root.Children[0]
	require.Equal(t, 1, nodeEntry.CountChildren())
	nsNode := nodeEntry.Children[0]
	require.GreaterOrEqual(t, nsNode.CountChildren(), 1)
	podNode := nsNode.Children[0]

	// Find PVC under the Pod and ensure the PV appears beneath it.
	var pvcNode *xray.TreeNode
	for _, c := range podNode.Children {
		if c.GVR == client.PvcGVR && c.ID == client.FQN("default", "web") {
			pvcNode = c
			break
		}
	}
	require.NotNil(t, pvcNode, "PVC node not found under pod")
	require.Equal(t, 1, pvcNode.CountChildren(), "PVC should have PV child")
	assert.Equal(t, client.PvGVR, pvcNode.Children[0].GVR)
	assert.Equal(t, client.FQN(client.ClusterScope, "pv-web"), pvcNode.Children[0].ID)
}

func TestNodeRender_ErrorOnBadInputType(t *testing.T) {
	var re xray.Node
	root := xray.NewTreeNode(client.NodeGVR, "nodes")
	ctx := context.WithValue(context.Background(), xray.KeyParent, root)
	ctx = context.WithValue(ctx, internal.KeyFactory, makeFactory())

	// Pass a non-unstructured type to trigger input type error.
	err := re.Render(ctx, "", &v1.Pod{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected Unstructured")
}

func TestNodeRender_ErrorWhenNoParentInContext(t *testing.T) {
	var re xray.Node
	o := load(t, "node")
	// No KeyParent set.
	ctx := context.WithValue(context.Background(), internal.KeyFactory, makeFactory())

	err := re.Render(ctx, "", o)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expecting a TreeNode")
}

func TestNodeRender_ErrorWhenNoFactoryInContext(t *testing.T) {
	var re xray.Node
	o := load(t, "node")
	root := xray.NewTreeNode(client.NodeGVR, "nodes")
	// No factory set.
	ctx := context.WithValue(context.Background(), xray.KeyParent, root)

	err := re.Render(ctx, "", o)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expecting a factory")
}

func TestNodeRender_NoPodsFound(t *testing.T) {
	// Factory has no pods listed for the node.
	f := makeFactory()

	var re xray.Node
	o := load(t, "node")
	root := xray.NewTreeNode(client.NodeGVR, "nodes")
	ctx := context.WithValue(context.Background(), xray.KeyParent, root)
	ctx = context.WithValue(ctx, internal.KeyFactory, f)

	require.NoError(t, re.Render(ctx, "", o))

	// Node is added under root, but it has no namespace/pod children.
	require.Equal(t, 1, root.CountChildren(), "expected one node entry under root")
	nodeEntry := root.Children[0]
	assert.Equal(t, 0, nodeEntry.CountChildren(), "expected no namespaces/pods under node")
}


