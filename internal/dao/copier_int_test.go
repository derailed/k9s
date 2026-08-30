// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func TestGenericCopy(t *testing.T) {
	uu := map[string]struct {
		path    string
		opts    *CopyOpts
		objs    []runtime.Object
		canI    bool
		e       string
		eRX     string
		err     string
		assertF func(*testing.T, dynamic.Interface)
	}{
		"happy": {
			path: "ns1/cm1",
			opts: &CopyOpts{Namespace: "ns2"},
			canI: true,
			e:    "ns2/cm1",
			assertF: func(t *testing.T, d dynamic.Interface) {
				o, err := d.Resource(client.CmGVR.GVR()).Namespace("ns2").
					Get(context.Background(), "cm1", metav1.GetOptions{})
				require.NoError(t, err)
				assert.Equal(t, "ns2", o.GetNamespace())
				// Server owned fields were dropped on the way over.
				assert.Empty(t, o.GetOwnerReferences())
				assert.Empty(t, o.GetLabels()["skip"])
				assert.Equal(t, map[string]string{"app": "fred"}, o.GetLabels())
			},
		},
		"rename": {
			path: "ns1/cm1",
			opts: &CopyOpts{Namespace: "ns2", Name: "cm2"},
			canI: true,
			e:    "ns2/cm2",
			assertF: func(t *testing.T, d dynamic.Interface) {
				o, err := d.Resource(client.CmGVR.GVR()).Namespace("ns2").
					Get(context.Background(), "cm2", metav1.GetOptions{})
				require.NoError(t, err)
				assert.Equal(t, "cm2", o.GetName())
			},
		},
		"rename-in-place": {
			path: "ns1/cm1",
			opts: &CopyOpts{Namespace: "ns1", Name: "cm1-copy"},
			canI: true,
			e:    "ns1/cm1-copy",
		},
		"same-ns-auto-suffix": {
			path: "ns1/cm1",
			opts: &CopyOpts{Namespace: "ns1"},
			canI: true,
			eRX:  `^ns1/cm1-[a-z0-9]{5}$`,
			assertF: func(t *testing.T, d dynamic.Interface) {
				// The source is left alone.
				o, err := d.Resource(client.CmGVR.GVR()).Namespace("ns1").
					Get(context.Background(), "cm1", metav1.GetOptions{})
				require.NoError(t, err)
				assert.Equal(t, "1234", string(o.GetUID()))
			},
		},
		"same-ns-explicit-name-auto-suffix": {
			path: "ns1/cm1",
			opts: &CopyOpts{Namespace: "ns1", Name: "cm1"},
			canI: true,
			eRX:  `^ns1/cm1-[a-z0-9]{5}$`,
		},
		"same-ns-overwrite-is-a-noop": {
			path: "ns1/cm1",
			opts: &CopyOpts{Namespace: "ns1", Overwrite: true},
			canI: true,
			err:  "source and target are identical: ns1/cm1",
		},
		"no-target-ns": {
			path: "ns1/cm1",
			opts: &CopyOpts{},
			canI: true,
			err:  "no target namespace specified",
		},
		"nil-opts": {
			path: "ns1/cm1",
			canI: true,
			err:  "no target namespace specified",
		},
		"cluster-scoped": {
			path: "-/cm1",
			opts: &CopyOpts{Namespace: "ns2"},
			canI: true,
			err:  "v1/configmaps is cluster scoped and cannot be copied to a namespace",
		},
		"unauthorized": {
			path: "ns1/cm1",
			opts: &CopyOpts{Namespace: "ns2"},
			canI: false,
			err:  `user is not authorized to create v1/configmaps in namespace "ns2"`,
		},
		"exists-no-overwrite": {
			path: "ns1/cm1",
			opts: &CopyOpts{Namespace: "ns2"},
			objs: []runtime.Object{makeCM("ns2", "cm1", "old")},
			canI: true,
			err:  `cm1 already exists in namespace "ns2". Enable overwrite to replace it`,
		},
		"exists-overwrite": {
			path: "ns1/cm1",
			opts: &CopyOpts{Namespace: "ns2", Overwrite: true},
			objs: []runtime.Object{makeCM("ns2", "cm1", "old")},
			canI: true,
			e:    "ns2/cm1",
			assertF: func(t *testing.T, d dynamic.Interface) {
				o, err := d.Resource(client.CmGVR.GVR()).Namespace("ns2").
					Get(context.Background(), "cm1", metav1.GetOptions{})
				require.NoError(t, err)
				v, ok, err := unstructured.NestedString(o.Object, "data", "k")
				require.NoError(t, err)
				assert.True(t, ok)
				assert.Equal(t, "src", v)
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			objs := append([]runtime.Object{makeSrcCM()}, u.objs...)
			dial := dynfake.NewSimpleDynamicClient(runtime.NewScheme(), objs...)

			var g Generic
			g.Init(newCopyFactory(dial, u.canI), client.CmGVR)

			fqn, err := g.Copy(context.Background(), u.path, u.opts)
			if u.err != "" {
				require.Error(t, err)
				assert.Equal(t, u.err, err.Error())
				return
			}
			require.NoError(t, err)
			if u.eRX != "" {
				assert.Regexp(t, u.eRX, fqn)
			} else {
				assert.Equal(t, u.e, fqn)
			}
			if u.assertF != nil {
				u.assertF(t, dial)
			}
		})
	}
}

// makeSrcCM builds a source configmap loaded up with server owned fields.
func makeSrcCM() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":            "cm1",
			"namespace":       "ns1",
			"uid":             "1234",
			"resourceVersion": "42",
			"labels":          map[string]any{"app": "fred"},
			"ownerReferences": []any{map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"name":       "p1",
				"uid":        "9999",
			}},
		},
		"data": map[string]any{"k": "src"},
	}}
}

func makeCM(ns, n, v string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":      n,
			"namespace": ns,
		},
		"data": map[string]any{"k": v},
	}}
}

// ----------------------------------------------------------------------------
// Helpers...

type copyFactory struct {
	conn *copyConn
}

func newCopyFactory(d dynamic.Interface, canI bool) Factory {
	return &copyFactory{conn: &copyConn{dial: d, canI: canI}}
}

func (f *copyFactory) Client() client.Connection { return f.conn }
func (*copyFactory) Get(*client.GVR, string, bool, labels.Selector) (runtime.Object, error) {
	return nil, nil
}

func (*copyFactory) List(*client.GVR, string, bool, labels.Selector) ([]runtime.Object, error) {
	return nil, nil
}

func (*copyFactory) ForResource(string, *client.GVR) (informers.GenericInformer, error) {
	return nil, nil
}

func (*copyFactory) CanForResource(string, *client.GVR, []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (*copyFactory) WaitForCacheSync()            {}
func (*copyFactory) Forwarders() watch.Forwarders { return nil }
func (*copyFactory) DeleteForwarder(string)       {}

type copyConn struct {
	dial dynamic.Interface
	canI bool
}

func (*copyConn) Config() *client.Config {
	return client.NewConfig(genericclioptions.NewConfigFlags(false))
}

func (c *copyConn) DynDial() (dynamic.Interface, error) { return c.dial, nil }

func (c *copyConn) CanI(string, *client.GVR, string, []string) (bool, error) {
	return c.canI, nil
}

func (*copyConn) Dial() (kubernetes.Interface, error)                      { return nil, nil }
func (*copyConn) DialLogs() (kubernetes.Interface, error)                  { return nil, nil }
func (*copyConn) ConnectionOK() bool                                       { return true }
func (*copyConn) SwitchContext(string) error                               { return nil }
func (*copyConn) CachedDiscovery() (*disk.CachedDiscoveryClient, error)    { return nil, nil }
func (*copyConn) RestConfig() (*restclient.Config, error)                  { return nil, nil }
func (*copyConn) MXDial() (*versioned.Clientset, error)                    { return nil, nil }
func (*copyConn) HasMetrics() bool                                         { return false }
func (*copyConn) CheckConnectivity() bool                                  { return true }
func (*copyConn) IsNamespaced(string) bool                                 { return true }
func (*copyConn) SupportsResource(string) bool                             { return true }
func (*copyConn) ValidNamespaces() ([]v1.Namespace, error)                 { return nil, nil }
func (*copyConn) SupportsRes(string, []string) (a string, b bool, e error) { return "", false, nil }
func (*copyConn) ServerVersion() (*version.Info, error)                    { return nil, nil }
func (*copyConn) CurrentNamespaceName() (string, error)                    { return "", nil }
func (*copyConn) ActiveContext() string                                    { return "" }
func (*copyConn) ActiveNamespace() string                                  { return "" }
func (*copyConn) IsValidNamespace(string) bool                             { return true }
func (*copyConn) ValidNamespaceNames() (client.NamespaceNames, error)      { return nil, nil }
func (*copyConn) IsActiveNamespace(string) bool                            { return false }

func TestSuffixName(t *testing.T) {
	uu := map[string]struct {
		n  string
		eL int
	}{
		"plain":     {n: "cm1", eL: len("cm1") + 1 + copySuffixLen},
		"truncated": {n: strings.Repeat("x", 200), eL: maxNameLen},
		"at-limit":  {n: strings.Repeat("x", maxNameLen), eL: maxNameLen},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			n := suffixName(u.n)
			assert.Len(t, n, u.eL)
			assert.LessOrEqual(t, len(n), maxNameLen)
			assert.Regexp(t, `-[a-z0-9]{5}$`, n)
			// Suffixes are random so two runs must not collide.
			assert.NotEqual(t, n, suffixName(u.n))
		})
	}
}

// TestCopyCreateSanitizes asserts a hand edited object still gets its status and
// server owned fields dropped on the way to the api-server.
func TestCopyCreateSanitizes(t *testing.T) {
	dial := dynfake.NewSimpleDynamicClient(runtime.NewScheme())

	var g Generic
	g.Init(newCopyFactory(dial, true), client.CmGVR)

	edited := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]any{
			"name":            "cm1",
			"namespace":       "ns2",
			"uid":             "pasted-back",
			"resourceVersion": "99",
			"ownerReferences": []any{map[string]any{"name": "dangling"}},
		},
		"data":   map[string]any{"k": "edited"},
		"status": map[string]any{"phase": "pasted-back"},
	}}

	fqn, err := g.CopyCreate(context.Background(), edited, false)
	require.NoError(t, err)
	assert.Equal(t, "ns2/cm1", fqn)

	o, err := dial.Resource(client.CmGVR.GVR()).Namespace("ns2").
		Get(context.Background(), "cm1", metav1.GetOptions{})
	require.NoError(t, err)

	_, ok, err := unstructured.NestedFieldNoCopy(o.Object, "status")
	require.NoError(t, err)
	assert.False(t, ok, "status must not survive a copy")
	assert.Empty(t, o.GetOwnerReferences())
	assert.Empty(t, string(o.GetUID()))
	// The user edit itself is preserved.
	v, ok, err := unstructured.NestedString(o.Object, "data", "k")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "edited", v)

	// The caller's object is left untouched.
	assert.Contains(t, edited.Object, "status")
}

func TestCopyCreateGuards(t *testing.T) {
	uu := map[string]struct {
		u    *unstructured.Unstructured
		canI bool
		err  string
	}{
		"nil":   {err: "no resource to copy"},
		"no-ns": {u: makeCM("", "cm1", "v"), canI: true, err: "no target namespace specified"},
		"no-name": {
			u: makeCM("ns2", "", "v"), canI: true, err: "no target name specified",
		},
		"unauthorized": {
			u:    makeCM("ns2", "cm1", "v"),
			canI: false,
			err:  `user is not authorized to create v1/configmaps in namespace "ns2"`,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			var g Generic
			g.Init(newCopyFactory(dynfake.NewSimpleDynamicClient(runtime.NewScheme()), u.canI), client.CmGVR)

			_, err := g.CopyCreate(context.Background(), u.u, false)
			require.Error(t, err)
			assert.Equal(t, u.err, err.Error())
		})
	}
}
