// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/xray"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodRender(t *testing.T) {
	uu := map[string]struct {
		file            string
		count, children int
		status          string
	}{
		"plain": {
			file:     "po",
			children: 1,
			count:    7,
			status:   xray.OkStatus,
		},
		"withInit": {
			file:     "init",
			children: 1,
			count:    7,
			status:   xray.OkStatus,
		},
		"cilium": {
			file:     "cilium",
			children: 1,
			count:    8,
			status:   xray.OkStatus,
		},
	}

	var re xray.Pod
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			o := load(t, u.file)
			root := xray.NewTreeNode(client.PodGVR, "pods")
			ctx := context.WithValue(context.Background(), xray.KeyParent, root)
			ctx = context.WithValue(ctx, internal.KeyFactory, makeFactory())

			require.NoError(t, re.Render(ctx, "", &render.PodWithMetrics{Raw: o}))
			assert.Equal(t, u.children, root.CountChildren())
			assert.Equal(t, u.count, root.Count(client.NoGVR))
		})
	}
}

func TestPodPVCExpandsToPV(t *testing.T) {
	// Arrange a factory that returns a bound PVC and its PV.
	f := makeFactory()
	f.rows = map[*client.GVR][]runtime.Object{
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

	var re xray.Pod
	o := load(t, "po")
	root := xray.NewTreeNode(client.PodGVR, "pods")
	ctx := context.WithValue(context.Background(), xray.KeyParent, root)
	ctx = context.WithValue(ctx, internal.KeyFactory, f)

	// Act
	require.NoError(t, re.Render(ctx, "", &render.PodWithMetrics{Raw: o}))

	// Assert: find PVC node under the rendered pod and ensure it has a PV child.
	require.Equal(t, 1, root.CountChildren())
	nsNode := root.Children[0]
	require.GreaterOrEqual(t, nsNode.CountChildren(), 1)
	podNode := nsNode.Children[0]

	var pvcNode *xray.TreeNode
	for _, c := range podNode.Children {
		if c.GVR == client.PvcGVR && c.ID == client.FQN("default", "web") {
			pvcNode = c
			break
		}
	}
	require.NotNil(t, pvcNode, "PVC node not found under pod")
	assert.Equal(t, 1, pvcNode.CountChildren(), "PVC should have PV child")
	assert.Equal(t, client.PvGVR, pvcNode.Children[0].GVR)
	assert.Equal(t, client.FQN(client.ClusterScope, "pv-web"), pvcNode.Children[0].ID)
}
