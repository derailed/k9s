package config_test

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/config"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

func TestNSValidate(t *testing.T) {
	setup(t)

	ns := config.NewNamespace()

	ksMock := NewMockKubeSettings()
	m.When(ksMock.NamespaceNames()).ThenReturn([]string{"ns1", "ns2", "default"}, nil)

	ns.Validate(ksMock)
	ksMock.VerifyWasCalledOnce()
	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"default"}, ns.Favorites)
}

func TestNSValidateMissing(t *testing.T) {
	setup(t)

	ns := config.NewNamespace()

	ksMock := NewMockKubeSettings()
	m.When(ksMock.NamespaceNames()).ThenReturn([]string{"ns1", "ns2"}, nil)
	ns.Validate(ksMock)

	ksMock.VerifyWasCalledOnce()
	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{}, ns.Favorites)
}

func TestNSValidateNoNS(t *testing.T) {
	setup(t)

	ns := config.NewNamespace()

	ksMock := NewMockKubeSettings()
	m.When(ksMock.NamespaceNames()).ThenReturn([]string{"ns1", "ns2"}, errors.New("boom"))
	ns.Validate(ksMock)

	ksMock.VerifyWasCalledOnce()
	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"default"}, ns.Favorites)
}

func TestNSSetActive(t *testing.T) {
	uu := []struct {
		ns  string
		fav []string
	}{
		{"all", []string{"all", "default"}},
		{"ns1", []string{"ns1", "all", "default"}},
		{"ns2", []string{"ns2", "ns1", "all", "default"}},
		{"ns3", []string{"ns3", "ns2", "ns1", "all", "default"}},
		{"ns4", []string{"ns4", "ns3", "ns2", "ns1", "all", "default"}},
	}

	ns := config.NewNamespace()
	for _, u := range uu {
		ns.SetActive(u.ns)
		assert.Equal(t, u.ns, ns.Active)
		assert.Equal(t, u.fav, ns.Favorites)
	}
}

func TestNSRmFavNS(t *testing.T) {
	ns := config.NewNamespace()
	uu := []struct {
		ns  string
		fav []string
	}{
		{"all", []string{"default", "kube-system"}},
		{"kube-system", []string{"default"}},
		{"blee", []string{"default"}},
	}

	for _, u := range uu {
		ns.SetActive(u.ns)
		assert.Equal(t, u.ns, ns.Active)
	}
}
