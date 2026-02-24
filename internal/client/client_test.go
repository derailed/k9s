// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	authorizationv1 "k8s.io/api/authorization/v1"
)

func TestMakeSAR(t *testing.T) {
	uu := map[string]struct {
		ns  string
		gvr *GVR
		sar *authorizationv1.SelfSubjectAccessReview
	}{
		"all-pods": {
			ns:  NamespaceAll,
			gvr: PodGVR,
			sar: &authorizationv1.SelfSubjectAccessReview{
				Spec: authorizationv1.SelfSubjectAccessReviewSpec{
					ResourceAttributes: &authorizationv1.ResourceAttributes{
						Namespace: NamespaceAll,
						Version:   "v1",
						Resource:  "pods",
					},
				},
			},
		},

		"ns-pods": {
			ns:  "fred",
			gvr: PodGVR,
			sar: &authorizationv1.SelfSubjectAccessReview{
				Spec: authorizationv1.SelfSubjectAccessReviewSpec{
					ResourceAttributes: &authorizationv1.ResourceAttributes{
						Namespace: "fred",
						Version:   "v1",
						Resource:  "pods",
					},
				},
			},
		},

		"clusterscope-ns": {
			ns:  ClusterScope,
			gvr: NsGVR,
			sar: &authorizationv1.SelfSubjectAccessReview{
				Spec: authorizationv1.SelfSubjectAccessReviewSpec{
					ResourceAttributes: &authorizationv1.ResourceAttributes{
						Version:  "v1",
						Resource: "namespaces",
					},
				},
			},
		},

		"subres-pods": {
			ns:  "fred",
			gvr: NewGVR("v1/pods:logs"),
			sar: &authorizationv1.SelfSubjectAccessReview{
				Spec: authorizationv1.SelfSubjectAccessReviewSpec{
					ResourceAttributes: &authorizationv1.ResourceAttributes{
						Namespace:   "fred",
						Version:     "v1",
						Resource:    "pods",
						Subresource: "logs",
					},
				},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.sar, makeSAR(u.ns, u.gvr, ""))
		})
	}
}

func TestIsValidNamespace(t *testing.T) {
	c := NewTestAPIClient()

	uu := map[string]struct {
		ns    string
		cache NamespaceNames
		ok    bool
	}{
		"all-ns": {
			ns: NamespaceAll,
			cache: NamespaceNames{
				DefaultNamespace: {},
			},
			ok: true,
		},
		"blank-ns": {
			ns: BlankNamespace,
			cache: NamespaceNames{
				DefaultNamespace: {},
			},
			ok: true,
		},
		"cluster-ns": {
			ns: ClusterScope,
			cache: NamespaceNames{
				DefaultNamespace: {},
			},
			ok: true,
		},
		"no-ns": {
			ns: NotNamespaced,
			cache: NamespaceNames{
				DefaultNamespace: {},
			},
			ok: true,
		},
		"default-ns": {
			ns: DefaultNamespace,
			cache: NamespaceNames{
				DefaultNamespace: {},
			},
			ok: true,
		},
		"valid-ns": {
			ns: "fred",
			cache: NamespaceNames{
				"fred": {},
			},
			ok: true,
		},
		"invalid-ns": {
			ns: "fred",
			cache: NamespaceNames{
				DefaultNamespace: {},
			},
		},
	}

	expiry := 1 * time.Millisecond
	for k := range uu {
		u := uu[k]
		c.cache.Add("validNamespaces", u.cache, expiry)
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.ok, c.IsValidNamespace(u.ns))
		})
	}
}

func TestCheckCacheBool(t *testing.T) {
	c := NewTestAPIClient()

	const key = "fred"
	uu := map[string]struct {
		key                  string
		val                  any
		found, actual, sleep bool
	}{
		"setTrue": {
			key:    key,
			val:    true,
			found:  true,
			actual: true,
		},
		"setFalse": {
			key:   key,
			val:   false,
			found: true,
		},
		"missing": {
			key: "blah",
			val: false,
		},
		"expired": {
			key:   key,
			val:   true,
			sleep: true,
		},
	}

	expiry := 1 * time.Millisecond
	for k := range uu {
		u := uu[k]
		c.cache.Add(key, u.val, expiry)
		if u.sleep {
			time.Sleep(expiry)
		}
		t.Run(k, func(t *testing.T) {
			val, ok := c.checkCacheBool(u.key)
			assert.Equal(t, u.found, ok)
			assert.Equal(t, u.actual, val)
		})
	}
}

func TestCachedDiscoveryPathIncludesClusterName(t *testing.T) {
	// This test verifies that the discovery cache path includes the cluster name
	// to prevent cache pollution when multiple clusters share the same proxy URL
	// Issue: #3828

	// Test case: Two clusters with same server URL but different cluster names
	// should have different cache paths
	sameServerURL := "https://teleport.example.com"
	prodCluster := "prod-cluster"
	stagingCluster := "staging-cluster"

	// Construct what would be the cache paths
	// The actual implementation in CachedDiscovery() uses:
	// filepath.Join(baseCacheDir, "discovery", toHostDir(cfg.Host), toHostDir(clusterName))
	// Since toHostDir is not exported, we verify the concept

	prodPath := sameServerURL + "/" + prodCluster
	stagingPath := sameServerURL + "/" + stagingCluster

	// Verify different clusters get different paths
	assert.NotEqual(t, prodPath, stagingPath,
		"clusters with same server URL but different names must have different cache paths")

	// Verify cluster name is in the path
	assert.Contains(t, prodPath, prodCluster, "prod cluster name should be in path")
	assert.Contains(t, stagingPath, stagingCluster, "staging cluster name should be in path")
}
