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

	ksMock := NewMockKubeSettings()
	m.When(ksMock.NamespaceNames()).ThenReturn(allNS, nil)

	ns := config.NewNamespace()
	for _, u := range uu {
		err := ns.SetActive(u.ns, ksMock)
		assert.Nil(t, err)
		assert.Equal(t, u.ns, ns.Active)
		assert.Equal(t, u.fav, ns.Favorites)
	}
}

func TestNSValidateRmFavs(t *testing.T) {
	allNS := []string{"default", "kube-system"}

	ksMock := NewMockKubeSettings()
	m.When(ksMock.NamespaceNames()).ThenReturn(allNS, nil)

	ns := config.NewNamespace()
	ns.Favorites = []string{"default", "fred", "blee"}

	ns.Validate(ksMock)
	assert.Equal(t, []string{"default"}, ns.Favorites)
}
