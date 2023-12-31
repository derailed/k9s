// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/k9s/internal/xray"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestCOConfigMapRefs(t *testing.T) {
	var re xray.Container

	root := xray.NewTreeNode("root", "root")
	ctx := context.WithValue(context.Background(), xray.KeyParent, root)
	ctx = context.WithValue(ctx, internal.KeyFactory, makeFactory())

	assert.Nil(t, re.Render(ctx, "", render.ContainerRes{Container: makeCMContainer("c1", false)}))
	assert.Equal(t, xray.MissingRefStatus, root.Children[0].Children[0].Extras[xray.StatusKey])
}

func TestCORefs(t *testing.T) {
	uu := map[string]struct {
		co             render.ContainerRes
		level1, level2 int
		e              string
	}{
		"cm_required": {
			co:     render.ContainerRes{Container: makeCMContainer("c1", false)},
			level1: 1,
			level2: 1,
			e:      xray.MissingRefStatus,
		},
		"cm_optional": {
			co:     render.ContainerRes{Container: makeCMContainer("c1", true)},
			level1: 1,
			level2: 1,
			e:      xray.OkStatus,
		},
		"cm_doubleRef": {
			co:     render.ContainerRes{Container: makeDoubleCMKeysContainer("c1", false)},
			level1: 1,
			level2: 1,
			e:      xray.MissingRefStatus,
		},
		"sec_required": {
			co:     render.ContainerRes{Container: makeSecContainer("c1", false)},
			level1: 1,
			level2: 1,
			e:      xray.MissingRefStatus,
		},
		"sec_optional": {
			co:     render.ContainerRes{Container: makeSecContainer("c1", true)},
			level1: 1,
			level2: 1,
			e:      xray.OkStatus,
		},
		"envFrom_optional": {
			co:     render.ContainerRes{Container: makeCMEnvFromContainer("c1", false)},
			level1: 1,
			level2: 2,
			e:      xray.MissingRefStatus,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			var re xray.Container
			root := xray.NewTreeNode("root", "root")
			ctx := context.WithValue(context.Background(), xray.KeyParent, root)
			ctx = context.WithValue(ctx, internal.KeyFactory, makeFactory())

			assert.Nil(t, re.Render(ctx, "", u.co))
			assert.Equal(t, u.level1, root.CountChildren())
			assert.Equal(t, u.level2, root.Children[0].CountChildren())
			assert.Equal(t, u.e, root.Children[0].Children[0].Extras[xray.StatusKey])
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeFactory() testFactory {
	return testFactory{}
}

type testFactory struct {
	rows map[string][]runtime.Object
}

var _ dao.Factory = testFactory{}

func (f testFactory) Client() client.Connection {
	return nil
}

func (f testFactory) Get(gvr, path string, wait bool, sel labels.Selector) (runtime.Object, error) {
	oo, ok := f.rows[gvr]
	if ok && len(oo) > 0 {
		return oo[0], nil
	}
	return nil, nil
}

func (f testFactory) List(gvr, ns string, wait bool, sel labels.Selector) ([]runtime.Object, error) {
	oo, ok := f.rows[gvr]
	if ok {
		return oo, nil
	}
	return nil, nil
}

func (f testFactory) ForResource(ns, gvr string) (informers.GenericInformer, error) {
	return nil, nil
}

func (f testFactory) CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (f testFactory) WaitForCacheSync() {}
func (f testFactory) Forwarders() watch.Forwarders {
	return nil
}
func (f testFactory) DeleteForwarder(string) {}

func makeCMEnvFromContainer(n string, optional bool) *v1.Container {
	return &v1.Container{
		Name: n,
		EnvFrom: []v1.EnvFromSource{
			{
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "cm1",
					},
					Optional: &optional,
				},
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "sec1",
					},
					Optional: &optional,
				},
			},
		},
	}
}

func makeCMContainer(n string, optional bool) *v1.Container {
	return &v1.Container{
		Name: n,
		Env: []v1.EnvVar{
			{
				Name: "e1",
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm1",
						},
						Key:      "k1",
						Optional: &optional,
					},
				},
			},
		},
	}
}

func makeSecContainer(n string, optional bool) *v1.Container {
	return &v1.Container{
		Name: n,
		Env: []v1.EnvVar{
			{
				Name: "e1",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "sec1",
						},
						Key:      "k1",
						Optional: &optional,
					},
				},
			},
		},
	}
}

func makeDoubleCMKeysContainer(n string, optional bool) *v1.Container {
	return &v1.Container{
		Name: n,
		Env: []v1.EnvVar{
			{
				Name: "e1",
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm1",
						},
						Key:      "k2",
						Optional: &optional,
					},
				},
			},
			{
				Name: "e2",
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "cm1",
						},
						Key:      "k1",
						Optional: &optional,
					},
				},
			},
		},
	}
}

func load(t *testing.T, n string) *unstructured.Unstructured {
	raw, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", n))
	assert.Nil(t, err)

	var o unstructured.Unstructured
	err = json.Unmarshal(raw, &o)
	assert.Nil(t, err)

	return &o
}
