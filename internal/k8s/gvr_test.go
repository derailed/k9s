package k8s_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestAsGR(t *testing.T) {
	uu := map[string]struct {
		gvr string
		e   schema.GroupVersion
	}{
		"full": {"apps/v1/deployments", schema.GroupVersion{Group: "apps", Version: "v1"}},
		"core": {"v1/pods", schema.GroupVersion{Group: "", Version: "v1"}},
		"bork": {"users", schema.GroupVersion{Group: "", Version: ""}},
	}

	for k := range uu {
   u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, k8s.GVR(u.gvr).AsGR())
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
			assert.Equal(t, u.e, k8s.NewGVR(u.g, u.v, u.r).String())
		})
	}
}

func TestToGVR(t *testing.T) {
	uu := map[string]struct {
		gv, r, e string
	}{
		"full": {"apps/v1", "deployments", "apps/v1/deployments"},
		"core": {"v1", "pods", "v1/pods"},
	}

	for k := range uu {
   u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, k8s.ToGVR(u.gv, u.r).String())
		})
	}
}

func TestResName(t *testing.T) {
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
			assert.Equal(t, u.e, k8s.GVR(u.gvr).ResName())
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
			assert.Equal(t, u.e, k8s.GVR(u.gvr).ToR())
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
			assert.Equal(t, u.e, k8s.GVR(u.gvr).ToG())
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
			assert.Equal(t, u.e, k8s.GVR(u.gvr).ToV())
		})
	}
}

func TestToStringer(t *testing.T) {
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
			assert.Equal(t, u.gvr, k8s.GVR(u.gvr).String())
		})
	}
}
