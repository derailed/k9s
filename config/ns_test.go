package config_test

import (
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/derailed/k9s/config"
	"github.com/stretchr/testify/assert"
)

func TestNSValidate(t *testing.T) {
	setup(t)

	ns := config.NewNamespace()

	ciMock := NewMockClusterInfo()
	m.When(ciMock.AllNamespacesOrDie()).ThenReturn([]string{"ns1", "ns2", "default"})

	ns.Validate(ciMock)
	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"all", "default"}, ns.Favorites)
}

func TestNSSetActive(t *testing.T) {
	uu := []struct {
		ns  string
		fav []string
	}{
		{"all", []string{"all", "default", "kube-system"}},
		{"ns1", []string{"ns1", "all", "default", "kube-system"}},
		{"ns2", []string{"ns2", "ns1", "all", "default", "kube-system"}},
		{"ns3", []string{"ns3", "ns2", "ns1", "all", "default", "kube-system"}},
		{"ns4", []string{"ns4", "ns3", "ns2", "ns1", "all", "default", "kube-system"}},
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
