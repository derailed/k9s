// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMetaFQN(t *testing.T) {
	uu := map[string]struct {
		meta metav1.ObjectMeta
		e    string
	}{
		"empty": {
			e: "-/",
		},
		"full": {
			meta: metav1.ObjectMeta{Name: "blee", Namespace: "ns1"},
			e:    "ns1/blee",
		},
		"no-ns": {
			meta: metav1.ObjectMeta{Name: "blee"},
			e:    "-/blee",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.MetaFQN(&u.meta))
		})
	}
}

func TestCoFQN(t *testing.T) {
	uu := map[string]struct {
		meta metav1.ObjectMeta
		co   string
		e    string
	}{
		"empty": {
			e: "-/:",
		},
		"full": {
			meta: metav1.ObjectMeta{Name: "blee", Namespace: "ns1"},
			co:   "fred",
			e:    "ns1/blee:fred",
		},
		"no-co": {
			meta: metav1.ObjectMeta{Name: "blee", Namespace: "ns1"},
			e:    "ns1/blee:",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.CoFQN(&u.meta, u.co))
		})
	}
}

func TestIsClusterScoped(t *testing.T) {
	uu := map[string]struct {
		ns string
		e  bool
	}{
		"empty": {},
		"all": {
			ns: client.NamespaceAll,
		},
		"none": {
			ns: client.BlankNamespace,
		},
		"custom": {
			ns: "fred",
		},
		"scoped": {
			ns: "-",
			e:  true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.IsClusterScoped(u.ns))
		})
	}
}

func TestIsNamespaced(t *testing.T) {
	uu := map[string]struct {
		ns string
		e  bool
	}{
		"empty": {},
		"all": {
			ns: client.NamespaceAll,
		},
		"cluster": {
			ns: client.ClusterScope,
		},
		"none": {
			ns: client.BlankNamespace,
		},
		"custom": {
			ns: "fred",
			e:  true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.IsNamespaced(u.ns))
		})
	}
}

func TestIsAllNamespaces(t *testing.T) {
	uu := map[string]struct {
		ns string
		e  bool
	}{
		"empty": {
			e: true,
		},
		"all": {
			ns: client.NamespaceAll,
			e:  true,
		},
		"none": {
			ns: client.BlankNamespace,
			e:  true,
		},
		"custom": {
			ns: "fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.IsAllNamespaces(u.ns))
		})
	}
}

func TestIsAllNamespace(t *testing.T) {
	uu := map[string]struct {
		ns string
		e  bool
	}{
		"empty": {},
		"all": {
			ns: client.NamespaceAll,
			e:  true,
		},
		"custom": {
			ns: "fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.IsAllNamespace(u.ns))
		})
	}
}

func TestCleanseNamespace(t *testing.T) {
	uu := map[string]struct {
		ns, e string
	}{
		"empty": {},
		"all": {
			ns: client.NamespaceAll,
			e:  client.BlankNamespace,
		},
		"custom": {
			ns: "fred",
			e:  "fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.CleanseNamespace(u.ns))
		})
	}
}

func TestNamespaced(t *testing.T) {
	uu := []struct {
		p, ns, n string
	}{
		{"fred/blee", "fred", "blee"},
		{"blee", "", "blee"},
	}

	for _, u := range uu {
		ns, n := client.Namespaced(u.p)
		assert.Equal(t, u.ns, ns)
		assert.Equal(t, u.n, n)
	}
}

func TestFQN(t *testing.T) {
	uu := []struct {
		ns, n string
		e     string
	}{
		{"fred", "blee", "fred/blee"},
		{"", "blee", "blee"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, client.FQN(u.ns, u.n))
	}
}

func TestIsMultiNamespace(t *testing.T) {
	uu := map[string]struct {
		ns string
		e  bool
	}{
		"empty":          {ns: client.BlankNamespace},
		"single":         {ns: "ns1"},
		"all":            {ns: client.NamespaceAll},
		"cluster":        {ns: client.ClusterScope},
		"multi":          {ns: "ns1,ns2", e: true},
		"trailing-comma": {ns: "ns1,", e: true},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.IsMultiNamespace(u.ns))
		})
	}
}

func TestNamespaces(t *testing.T) {
	uu := map[string]struct {
		ns string
		e  []string
	}{
		"single":  {ns: "ns1", e: []string{"ns1"}},
		"multi":   {ns: "ns1,ns2", e: []string{"ns1", "ns2"}},
		"spaces":  {ns: " ns1 , ns2 ", e: []string{"ns1", "ns2"}},
		"empties": {ns: "ns1,,ns2,", e: []string{"ns1", "ns2"}},
		"blank":   {ns: client.BlankNamespace, e: []string{}},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.Namespaces(u.ns))
		})
	}
}

func TestNormalizeNamespaces(t *testing.T) {
	uu := map[string]struct {
		ns string
		e  string
	}{
		"single":     {ns: "ns1", e: "ns1"},
		"all":        {ns: client.NamespaceAll, e: client.NamespaceAll},
		"multi":      {ns: "ns1,ns2", e: "ns1,ns2"},
		"dedup":      {ns: "ns1,ns2,ns1", e: "ns1,ns2"},
		"trim-empty": {ns: " ns1 , , ns2 ,", e: "ns1,ns2"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.NormalizeNamespaces(u.ns))
		})
	}
}
