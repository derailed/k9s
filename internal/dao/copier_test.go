// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSanitizeForCopyStripsServerFields(t *testing.T) {
	u := unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":              "cm1",
			"namespace":         "ns1",
			"uid":               "1234",
			"resourceVersion":   "42",
			"generation":        int64(3),
			"creationTimestamp": "2026-01-01T00:00:00Z",
			"selfLink":          "/api/v1/blee",
			"finalizers":        []any{"blee"},
			"managedFields":     []any{map[string]any{"manager": "kubectl"}},
			"ownerReferences":   []any{map[string]any{"name": "duh"}},
			"labels":            map[string]any{"app": "fred"},
		},
		"data":   map[string]any{"k": "v"},
		"status": map[string]any{"phase": "Active"},
	}}

	sanitizeForCopy(&u, client.CmGVR)

	for _, f := range [][]string{
		{"metadata", "uid"},
		{"metadata", "resourceVersion"},
		{"metadata", "generation"},
		{"metadata", "creationTimestamp"},
		{"metadata", "selfLink"},
		{"metadata", "finalizers"},
		{"metadata", "managedFields"},
		{"metadata", "ownerReferences"},
		{"status"},
	} {
		_, ok, err := unstructured.NestedFieldNoCopy(u.Object, f...)
		require.NoError(t, err)
		assert.False(t, ok, "expected %v to be pruned", f)
	}

	// User owned bits survive.
	assert.Equal(t, "cm1", u.GetName())
	assert.Equal(t, map[string]string{"app": "fred"}, u.GetLabels())
	d, ok, err := unstructured.NestedStringMap(u.Object, "data")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, map[string]string{"k": "v"}, d)
}

func TestSanitizeForCopyAnnotations(t *testing.T) {
	uu := map[string]struct {
		in, e map[string]any
		keep  bool
	}{
		"prunes-stale": {
			in: map[string]any{
				"kubectl.kubernetes.io/last-applied-configuration": "{}",
				"deployment.kubernetes.io/revision":                "3",
				"blee":                                             "duh",
			},
			e:    map[string]any{"blee": "duh"},
			keep: true,
		},
		"drops-empty-map": {
			in: map[string]any{
				"kubectl.kubernetes.io/last-applied-configuration": "{}",
			},
			keep: false,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			o := unstructured.Unstructured{Object: map[string]any{
				"metadata": map[string]any{
					"name":        "cm1",
					"annotations": u.in,
				},
			}}
			sanitizeForCopy(&o, client.CmGVR)

			aa, ok, err := unstructured.NestedFieldNoCopy(o.Object, "metadata", "annotations")
			require.NoError(t, err)
			assert.Equal(t, u.keep, ok)
			if u.keep {
				assert.Equal(t, u.e, aa)
			}
		})
	}
}

func TestSanitizeForCopySvc(t *testing.T) {
	u := unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{"name": "svc1"},
		"spec": map[string]any{
			"clusterIP":      "10.0.0.1",
			"clusterIPs":     []any{"10.0.0.1"},
			"ipFamilies":     []any{"IPv4"},
			"ipFamilyPolicy": "SingleStack",
			"type":           "NodePort",
			"selector":       map[string]any{"app": "fred"},
			"ports": []any{
				map[string]any{"name": "http", "port": int64(80), "nodePort": int64(31000)},
				map[string]any{"name": "https", "port": int64(443), "nodePort": int64(31001)},
			},
		},
	}}

	sanitizeForCopy(&u, client.SvcGVR)

	for _, f := range []string{"clusterIP", "clusterIPs", "ipFamilies", "ipFamilyPolicy"} {
		_, ok, err := unstructured.NestedFieldNoCopy(u.Object, "spec", f)
		require.NoError(t, err)
		assert.False(t, ok, "expected spec.%s to be pruned", f)
	}

	pp, ok, err := unstructured.NestedSlice(u.Object, "spec", "ports")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Len(t, pp, 2)
	for _, p := range pp {
		m, isMap := p.(map[string]any)
		assert.True(t, isMap)
		assert.NotContains(t, m, "nodePort")
		assert.Contains(t, m, "port")
	}

	// Untouched spec bits survive.
	sel, ok, err := unstructured.NestedStringMap(u.Object, "spec", "selector")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, map[string]string{"app": "fred"}, sel)
}

func TestSanitizeForCopyJob(t *testing.T) {
	u := unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{
			"name": "job1",
			"labels": map[string]any{
				"controller-uid":                     "1234",
				"job-name":                           "job1",
				"batch.kubernetes.io/controller-uid": "1234",
				"batch.kubernetes.io/job-name":       "job1",
				"app":                                "fred",
			},
		},
		"spec": map[string]any{
			"manualSelector": true,
			"selector": map[string]any{
				"matchLabels": map[string]any{"controller-uid": "1234"},
			},
			"template": map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"controller-uid": "1234",
						"job-name":       "job1",
					},
				},
			},
		},
	}}

	sanitizeForCopy(&u, client.JobGVR)

	for _, f := range []string{"selector", "manualSelector"} {
		_, ok, err := unstructured.NestedFieldNoCopy(u.Object, "spec", f)
		require.NoError(t, err)
		assert.False(t, ok, "expected spec.%s to be pruned", f)
	}

	assert.Equal(t, map[string]string{"app": "fred"}, u.GetLabels())

	// All pod template labels were tracking labels, so the map goes away.
	_, ok, err := unstructured.NestedFieldNoCopy(u.Object, "spec", "template", "metadata", "labels")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestSanitizeForCopyBoundFields(t *testing.T) {
	uu := map[string]struct {
		gvr   *client.GVR
		obj   map[string]any
		gone  [][]string
		stays [][]string
	}{
		"pod": {
			gvr: client.PodGVR,
			obj: map[string]any{
				"metadata": map[string]any{"name": "p1"},
				"spec":     map[string]any{"nodeName": "n1", "restartPolicy": "Always"},
			},
			gone:  [][]string{{"spec", "nodeName"}},
			stays: [][]string{{"spec", "restartPolicy"}},
		},
		"pvc": {
			gvr: client.PvcGVR,
			obj: map[string]any{
				"metadata": map[string]any{"name": "pvc1"},
				"spec":     map[string]any{"volumeName": "pv1", "storageClassName": "gp2"},
			},
			gone:  [][]string{{"spec", "volumeName"}},
			stays: [][]string{{"spec", "storageClassName"}},
		},
		"sa": {
			gvr: client.SaGVR,
			obj: map[string]any{
				"metadata": map[string]any{"name": "sa1"},
				"secrets":  []any{map[string]any{"name": "sa1-token"}},
			},
			gone: [][]string{{"secrets"}},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			o := unstructured.Unstructured{Object: u.obj}
			sanitizeForCopy(&o, u.gvr)

			for _, f := range u.gone {
				_, ok, err := unstructured.NestedFieldNoCopy(o.Object, f...)
				require.NoError(t, err)
				assert.False(t, ok, "expected %v to be pruned", f)
			}
			for _, f := range u.stays {
				_, ok, err := unstructured.NestedFieldNoCopy(o.Object, f...)
				require.NoError(t, err)
				assert.True(t, ok, "expected %v to survive", f)
			}
		})
	}
}
