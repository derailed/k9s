// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client_test

import (
	"path"
	"sort"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGVRSort(t *testing.T) {
	gg := client.GVRs{
		client.NewGVR("v1/pods"),
		client.NewGVR("v1/services"),
		client.NewGVR("apps/v1/deployments"),
	}
	sort.Sort(gg)
	assert.Equal(t, client.GVRs{
		client.NewGVR("v1/pods"),
		client.NewGVR("v1/services"),
		client.NewGVR("apps/v1/deployments"),
	}, gg)
}

func TestGVRCan(t *testing.T) {
	uu := map[string]struct {
		vv []string
		v  string
		e  bool
	}{
		"describe":  {[]string{"get"}, "describe", true},
		"view":      {[]string{"get", "list", "watch"}, "view", true},
		"delete":    {[]string{"delete", "list", "watch"}, "delete", true},
		"no_delete": {[]string{"get", "list", "watch"}, "delete", false},
		"edit":      {[]string{"path", "update", "watch"}, "edit", true},
		"no_edit":   {[]string{"get", "list", "watch"}, "edit", false},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.Can(u.vv, u.v))
		})
	}
}

func TestGVR(t *testing.T) {
	uu := map[string]struct {
		gvr string
		e   schema.GroupVersionResource
	}{
		"full": {"apps/v1/deployments", schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}},
		"core": {"v1/pods", schema.GroupVersionResource{Version: "v1", Resource: "pods"}},
		"bork": {"users", schema.GroupVersionResource{Resource: "users"}},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.NewGVR(u.gvr).GVR())
		})
	}
}

func TestAsGV(t *testing.T) {
	uu := map[string]struct {
		gvr string
		e   schema.GroupVersion
	}{
		"full": {"apps/v1/deployments", schema.GroupVersion{Group: "apps", Version: "v1"}},
		"core": {"v1/pods", schema.GroupVersion{Version: "v1"}},
		"bork": {"users", schema.GroupVersion{}},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.NewGVR(u.gvr).GV())
		})
	}
}

func TestNewGVR(t *testing.T) {
	uu := map[string]struct {
		g, v, r string
		e       string
	}{
		"full": {"apps", "v1", "deployments", "apps/v1/deployments"},
		"core": {"", "v1", "pods", "v1/pods"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.NewGVR(path.Join(u.g, u.v, u.r)).String())
		})
	}
}

func TestGVRAsResourceName(t *testing.T) {
	uu := map[string]struct {
		gvr string
		e   string
	}{
		"full":  {"apps/v1/deployments", "deployments.v1.apps"},
		"core":  {"v1/pods", "pods.v1."},
		"k9s":   {"users", "users.."},
		"empty": {"", ".."},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.NewGVR(u.gvr).AsResourceName())
		})
	}
}

func TestToR(t *testing.T) {
	uu := map[string]struct {
		gvr string
		e   string
	}{
		"full":  {"apps/v1/deployments", "deployments"},
		"core":  {"v1/pods", "pods"},
		"k9s":   {"users", "users"},
		"empty": {"", ""},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.NewGVR(u.gvr).R())
		})
	}
}

func TestToG(t *testing.T) {
	uu := map[string]struct {
		gvr string
		e   string
	}{
		"full":  {"apps/v1/deployments", "apps"},
		"core":  {"v1/pods", ""},
		"k9s":   {"users", ""},
		"empty": {"", ""},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.NewGVR(u.gvr).G())
		})
	}
}

func TestToV(t *testing.T) {
	uu := map[string]struct {
		gvr string
		e   string
	}{
		"full":  {"apps/v1/deployments", "v1"},
		"core":  {"v1beta1/pods", "v1beta1"},
		"k9s":   {"users", ""},
		"empty": {"", ""},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, client.NewGVR(u.gvr).V())
		})
	}
}

func TestToString(t *testing.T) {
	uu := map[string]struct {
		gvr string
	}{
		"full":  {"apps/v1/deployments"},
		"core":  {"v1beta1/pods"},
		"k9s":   {"users"},
		"empty": {""},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.gvr, client.NewGVR(u.gvr).String())
		})
	}
}
