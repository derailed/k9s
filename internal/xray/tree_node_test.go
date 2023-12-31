// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/xray"
	"github.com/stretchr/testify/assert"
)

func TestTreeNodeCount(t *testing.T) {
	uu := map[string]struct {
		root *xray.TreeNode
		e    int
	}{
		"simple": {
			root: root1(),
			e:    3,
		},
		"complex": {
			root: root3(),
			e:    26,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.root.Count(""))
		})
	}
}

func TestTreeNodeFilter(t *testing.T) {
	uu := map[string]struct {
		q       string
		root, e *xray.TreeNode
	}{
		"filter_simple": {
			root: root1(),
			e:    diff1(),
			q:    "c1",
		},
		"filter_complex": {
			root: root2(),
			e:    diff2(),
			q:    "c2",
		},
		"filter_no_match": {
			root: root2(),
			e:    nil,
			q:    "bozo",
		},
		"filter_all_match": {
			root: root2(),
			e:    root2(),
			q:    "",
		},
		"filter_complex1": {
			root: root3(),
			e:    diff3(),
			q:    "coredns",
		},
	}

	rx := func(q, path string) bool {
		rx := regexp.MustCompile(`(?i)` + q)

		tokens := strings.Split(path, "::")
		for _, t := range tokens {
			if rx.MatchString(t) {
				return true
			}
		}
		return false
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			filtered := u.root.Filter(u.q, rx)
			assert.Equal(t, u.e, filtered)
		})
	}
}

func TestTreeNodeHydrate(t *testing.T) {
	threeOK := []string{"ok", "ok", "ok"}
	fiveOK := append(threeOK, "ok", "ok")

	uu := map[string]struct {
		spec []xray.NodeSpec
		e    *xray.TreeNode
	}{
		"flat_simple": {
			spec: []xray.NodeSpec{
				{
					GVRs:     []string{"containers", "v1/pods"},
					Paths:    []string{"c1", "default/p1"},
					Statuses: threeOK,
				},
				{
					GVRs:     []string{"containers", "v1/pods"},
					Paths:    []string{"c2", "default/p1"},
					Statuses: threeOK,
				},
			},
			e: root1(),
		},
		"flat_complex": {
			spec: []xray.NodeSpec{
				{
					GVRs:     []string{"v1/secrets", "containers", "v1/pods"},
					Paths:    []string{"s1", "c1", "default/p1"},
					Statuses: threeOK,
				},
				{
					GVRs:     []string{"v1/secrets", "containers", "v1/pods"},
					Paths:    []string{"s2", "c2", "default/p1"},
					Statuses: threeOK,
				},
			},
			e: root2(),
		},
		"complex1": {
			spec: []xray.NodeSpec{
				{
					GVRs:     []string{"v1/secrets", "v1/pods", "apps/v1/deployments", "v1/namespaces", "apps/v1/deployments"},
					Paths:    []string{"default/default-token-rr22g", "default/nginx-6b866d578b-c6tcn", "default/nginx", "-/default", "deployments"},
					Statuses: fiveOK,
				},
				{
					GVRs:     []string{"v1/configmaps", "v1/pods", "apps/v1/deployments", "v1/namespaces", "apps/v1/deployments"},
					Paths:    []string{"kube-system/coredns", "kube-system/coredns-6955765f44-89q2p", "kube-system/coredns", "-/kube-system", "deployments"},
					Statuses: fiveOK,
				},
				{
					GVRs:     []string{"v1/secrets", "v1/pods", "apps/v1/deployments", "v1/namespaces", "apps/v1/deployments"},
					Paths:    []string{"kube-system/coredns-token-5cq9j", "kube-system/coredns-6955765f44-89q2p", "kube-system/coredns", "-/kube-system", "deployments"},
					Statuses: fiveOK,
				},
				{
					GVRs:     []string{"v1/configmaps", "v1/pods", "apps/v1/deployments", "v1/namespaces", "apps/v1/deployments"},
					Paths:    []string{"kube-system/coredns", "kube-system/coredns-6955765f44-r9j9t", "kube-system/coredns", "-/kube-system", "deployments"},
					Statuses: fiveOK,
				},
				{
					GVRs:     []string{"v1/secrets", "v1/pods", "apps/v1/deployments", "v1/namespaces", "apps/v1/deployments"},
					Paths:    []string{"kube-system/coredns-token-5cq9j", "kube-system/coredns-6955765f44-r9j9t", "kube-system/coredns", "-/kube-system", "deployments"},
					Statuses: fiveOK,
				},
				{
					GVRs:     []string{"v1/secrets", "v1/pods", "apps/v1/deployments", "v1/namespaces", "apps/v1/deployments"},
					Paths:    []string{"kube-system/default-token-thzt8", "kube-system/metrics-server-6754dbc9df-88bk4", "kube-system/metrics-server", "-/kube-system", "deployments"},
					Statuses: fiveOK,
				},
				{
					GVRs:     []string{"v1/secrets", "v1/pods", "apps/v1/deployments", "v1/namespaces", "apps/v1/deployments"},
					Paths:    []string{"kube-system/nginx-ingress-token-kff5q", "kube-system/nginx-ingress-controller-6fc5bcc8c9-cwp55", "kube-system/nginx-ingress-controller", "-/kube-system", "deployments"},
					Statuses: fiveOK,
				},
				{
					GVRs:     []string{"v1/secrets", "v1/pods", "apps/v1/deployments", "v1/namespaces", "apps/v1/deployments"},
					Paths:    []string{"kubernetes-dashboard/kubernetes-dashboard-token-d6rt4", "kubernetes-dashboard/dashboard-metrics-scraper-7b64584c5c-c7b56", "kubernetes-dashboard/dashboard-metrics-scraper", "-/kubernetes-dashboard", "deployments"},
					Statuses: fiveOK,
				},
				{
					GVRs:     []string{"v1/secrets", "v1/pods", "apps/v1/deployments", "v1/namespaces", "apps/v1/deployments"},
					Paths:    []string{"kubernetes-dashboard/kubernetes-dashboard-token-d6rt4", "kubernetes-dashboard/kubernetes-dashboard-79d9cd965-b4c7d", "kubernetes-dashboard/kubernetes-dashboard", "-/kubernetes-dashboard", "deployments"},
					Statuses: fiveOK,
				},
			},
			e: root3(),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			root := xray.Hydrate(u.spec)
			assert.Equal(t, u.e.Flatten(), root.Flatten())
		})
	}
}

func TestTreeNodeFlatten(t *testing.T) {
	uu := map[string]struct {
		root *xray.TreeNode
		e    []xray.NodeSpec
	}{
		"flat_simple": {
			root: root1(),
			e: []xray.NodeSpec{
				{
					GVRs:     []string{"containers", "v1/pods"},
					Paths:    []string{"c1", "default/p1"},
					Statuses: []string{"ok", "ok"},
				},
				{
					GVRs:     []string{"containers", "v1/pods"},
					Paths:    []string{"c2", "default/p1"},
					Statuses: []string{"ok", "ok"},
				},
			},
		},
		"flat_complex": {
			root: root2(),
			e: []xray.NodeSpec{
				{
					GVRs:     []string{"v1/secrets", "containers", "v1/pods"},
					Paths:    []string{"s1", "c1", "default/p1"},
					Statuses: []string{"ok", "ok", "ok"},
				},
				{
					GVRs:     []string{"v1/secrets", "containers", "v1/pods"},
					Paths:    []string{"s2", "c2", "default/p1"},
					Statuses: []string{"ok", "ok", "ok"},
				},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			flat := u.root.Flatten()
			assert.Equal(t, u.e, flat)
		})
	}
}

func TestTreeNodeDiff(t *testing.T) {
	uu := map[string]struct {
		n1, n2 *xray.TreeNode
		e      bool
	}{
		"blank": {
			n1: &xray.TreeNode{},
			n2: &xray.TreeNode{},
		},
		"same": {
			n1: xray.NewTreeNode("v1/pods", "default/p1"),
			n2: xray.NewTreeNode("v1/pods", "default/p1"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.n1.Diff(u.n2))
		})
	}
}

func TestTreeNodeClone(t *testing.T) {
	n := xray.NewTreeNode("v1/pods", "default/p1")
	c1 := xray.NewTreeNode("containers", "c1")
	n.Add(c1)

	c := n.ShallowClone()
	assert.Equal(t, n.GVR, c.GVR)
}

func TestTreeNodeRoot(t *testing.T) {
	n := xray.NewTreeNode("v1/pods", "default/p1")
	c1 := xray.NewTreeNode("containers", "c1")
	c2 := xray.NewTreeNode("containers", "c2")
	n.Add(c1)
	n.Add(c2)

	assert.Equal(t, 2, n.CountChildren())
	assert.Equal(t, n, n.Root())
	assert.True(t, n.IsRoot())
	assert.False(t, n.IsLeaf())
	assert.Equal(t, n, c1.Root())
	assert.False(t, c1.IsRoot())
	assert.Equal(t, n, c2.Root())
	assert.True(t, c1.IsLeaf())
}

func TestTreeNodeLevel(t *testing.T) {
	n := xray.NewTreeNode("v1/pods", "default/p1")
	c1 := xray.NewTreeNode("containers", "c1")
	c2 := xray.NewTreeNode("containers", "c2")
	n.Add(c1)
	n.Add(c2)

	assert.Equal(t, 0, n.Level())
	assert.Equal(t, 1, c1.Level())
	assert.Equal(t, 1, c2.Level())
}

func TestTreeNodeMaxDepth(t *testing.T) {
	n := xray.NewTreeNode("v1/pods", "default/p1")
	c1 := xray.NewTreeNode("containers", "c1")
	c2 := xray.NewTreeNode("containers", "c2")
	n.Add(c1)
	n.Add(c2)

	assert.Equal(t, 1, n.MaxDepth(0))
}

// ----------------------------------------------------------------------------
// Helpers...

func root1() *xray.TreeNode {
	n := xray.NewTreeNode("v1/pods", "default/p1")
	c1 := xray.NewTreeNode("containers", "c1")
	c2 := xray.NewTreeNode("containers", "c2")
	n.Add(c1)
	n.Add(c2)

	return n
}

func diff1() *xray.TreeNode {
	n := xray.NewTreeNode("v1/pods", "default/p1")
	c1 := xray.NewTreeNode("containers", "c1")
	n.Add(c1)

	return n
}

func root2() *xray.TreeNode {
	c1 := xray.NewTreeNode("containers", "c1")
	s1 := xray.NewTreeNode("v1/secrets", "s1")
	c1.Add(s1)

	c2 := xray.NewTreeNode("containers", "c2")
	s2 := xray.NewTreeNode("v1/secrets", "s2")
	c2.Add(s2)

	n := xray.NewTreeNode("v1/pods", "default/p1")
	n.Add(c1)
	n.Add(c2)

	return n
}

func diff2() *xray.TreeNode {
	n := xray.NewTreeNode("v1/pods", "default/p1")
	c1 := xray.NewTreeNode("containers", "c2")
	n.Add(c1)

	s1 := xray.NewTreeNode("v1/secrets", "s2")
	c1.Add(s1)

	return n
}

func root3() *xray.TreeNode {
	n := xray.NewTreeNode("apps/v1/deployments", "deployments")

	ns1 := xray.NewTreeNode("v1/namespaces", "-/default")
	n.Add(ns1)
	{
		d1 := xray.NewTreeNode("apps/v1/deployments", "default/nginx")
		ns1.Add(d1)
		{
			p1 := xray.NewTreeNode("v1/pods", "default/nginx-6b866d578b-c6tcn")
			d1.Add(p1)
			{
				s1 := xray.NewTreeNode("v1/secrets", "default/default-token-rr22g")
				p1.Add(s1)
			}
		}
	}

	ns2 := xray.NewTreeNode("v1/namespaces", "-/kube-system")
	n.Add(ns2)
	{
		d2 := xray.NewTreeNode("apps/v1/deployments", "kube-system/coredns")
		ns2.Add(d2)
		{
			p2 := xray.NewTreeNode("v1/pods", "kube-system/coredns-6955765f44-89q2p")
			d2.Add(p2)
			{
				c1 := xray.NewTreeNode("v1/configmaps", "kube-system/coredns")
				p2.Add(c1)
				s2 := xray.NewTreeNode("v1/secrets", "kube-system/coredns-token-5cq9j")
				p2.Add(s2)
			}
			p3 := xray.NewTreeNode("v1/pods", "kube-system/coredns-6955765f44-r9j9t")
			d2.Add(p3)
			{
				c2 := xray.NewTreeNode("v1/configmaps", "kube-system/coredns")
				p3.Add(c2)
				s3 := xray.NewTreeNode("v1/secrets", "kube-system/coredns-token-5cq9j")
				p3.Add(s3)
			}
		}
		d3 := xray.NewTreeNode("apps/v1/deployments", "kube-system/metrics-server")
		ns2.Add(d3)
		{
			p3 := xray.NewTreeNode("v1/pods", "kube-system/metrics-server-6754dbc9df-88bk4")
			d3.Add(p3)
			{
				s4 := xray.NewTreeNode("v1/secrets", "kube-system/default-token-thzt8")
				p3.Add(s4)
			}
		}
		d4 := xray.NewTreeNode("apps/v1/deployments", "kube-system/nginx-ingress-controller")
		ns2.Add(d4)
		{
			p4 := xray.NewTreeNode("v1/pods", "kube-system/nginx-ingress-controller-6fc5bcc8c9-cwp55")
			d4.Add(p4)
			{
				s5 := xray.NewTreeNode("v1/secrets", "kube-system/nginx-ingress-token-kff5q")
				p4.Add(s5)
			}
		}
	}

	ns3 := xray.NewTreeNode("v1/namespaces", "-/kubernetes-dashboard")
	n.Add(ns3)
	{
		d5 := xray.NewTreeNode("apps/v1/deployments", "kubernetes-dashboard/dashboard-metrics-scraper")
		ns3.Add(d5)
		{
			p5 := xray.NewTreeNode("v1/pods", "kubernetes-dashboard/dashboard-metrics-scraper-7b64584c5c-c7b56")
			d5.Add(p5)
			{
				s6 := xray.NewTreeNode("v1/secrets", "kubernetes-dashboard/kubernetes-dashboard-token-d6rt4")
				p5.Add(s6)
			}
		}
		d6 := xray.NewTreeNode("apps/v1/deployments", "kubernetes-dashboard/kubernetes-dashboard")
		ns3.Add(d6)
		{
			p6 := xray.NewTreeNode("v1/pods", "kubernetes-dashboard/kubernetes-dashboard-79d9cd965-b4c7d")
			d6.Add(p6)
			{
				s6 := xray.NewTreeNode("v1/secrets", "kubernetes-dashboard/kubernetes-dashboard-token-d6rt4")
				p6.Add(s6)
			}
		}
	}

	return n
}

func diff3() *xray.TreeNode {
	n := xray.NewTreeNode("apps/v1/deployments", "deployments")
	ns2 := xray.NewTreeNode("v1/namespaces", "-/kube-system")
	n.Add(ns2)
	{
		d2 := xray.NewTreeNode("apps/v1/deployments", "kube-system/coredns")
		ns2.Add(d2)
		{
			p2 := xray.NewTreeNode("v1/pods", "kube-system/coredns-6955765f44-89q2p")
			d2.Add(p2)
			{
				c1 := xray.NewTreeNode("v1/configmaps", "kube-system/coredns")
				p2.Add(c1)
				s2 := xray.NewTreeNode("v1/secrets", "kube-system/coredns-token-5cq9j")
				p2.Add(s2)
			}
			p3 := xray.NewTreeNode("v1/pods", "kube-system/coredns-6955765f44-r9j9t")
			d2.Add(p3)
			{
				c2 := xray.NewTreeNode("v1/configmaps", "kube-system/coredns")
				p3.Add(c2)
				s3 := xray.NewTreeNode("v1/secrets", "kube-system/coredns-token-5cq9j")
				p3.Add(s3)
			}
		}
	}
	return n
}
