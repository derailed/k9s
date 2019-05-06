package config_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/config"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

func TestNSValidate(t *testing.T) {
	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)
	mk := NewMockKubeSettings()
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"ns1", "ns2", "default"})

	ns := config.NewNamespace()
	ns.Validate(mc, mk)

	mk.VerifyWasCalledOnce()
	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"default"}, ns.Favorites)
}

func TestNSValidateMissing(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatalf("Expected panic on non existing namespace")
		}
	}()

	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)
	mk := NewMockKubeSettings()
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"ns1", "ns2"})

	ns := config.NewNamespace()
	ns.Validate(mc, mk)

	mk.VerifyWasCalledOnce()
	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{}, ns.Favorites)
}

func TestNSValidateNoNS(t *testing.T) {
	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), fmt.Errorf("Crap!"))
	mk := NewMockKubeSettings()
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"ns1", "ns2"})

	ns := config.NewNamespace()
	ns.Validate(mc, mk)

	mk.VerifyWasCalledOnce()
	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"default"}, ns.Favorites)
}

func TestNSSetActive(t *testing.T) {
	allNS := []string{"ns4", "ns3", "ns2", "ns1", "all", "default"}
	uu := []struct {
		ns  string
		fav []string
	}{
		{"all", []string{"all", "default"}},
		{"ns1", []string{"ns1", "all", "default"}},
		{"ns2", []string{"ns2", "ns1", "all", "default"}},
		{"ns3", []string{"ns3", "ns2", "ns1", "all", "default"}},
		{"ns4", allNS},
	}

	mk := NewMockKubeSettings()
	m.When(mk.NamespaceNames(namespaces())).ThenReturn(allNS)

	ns := config.NewNamespace()
	for _, u := range uu {
		err := ns.SetActive(u.ns, mk)

		assert.Nil(t, err)
		assert.Equal(t, u.ns, ns.Active)
		assert.Equal(t, u.fav, ns.Favorites)
	}
}

func TestNSValidateRmFavs(t *testing.T) {
	allNS := []string{"default", "kube-system"}

	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)

	mk := NewMockKubeSettings()
	m.When(mk.NamespaceNames(namespaces())).ThenReturn(allNS)

	ns := config.NewNamespace()
	ns.Favorites = []string{"default", "fred", "blee"}
	ns.Validate(mc, mk)

	assert.Equal(t, []string{"default"}, ns.Favorites)
}
