package views

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	k9sCfg.load("../test_assets/k9s1.yml")

	assert.Equal(t, 10, k9sCfg.K9s.RefreshRate)
	assert.Equal(t, "fred", k9sCfg.K9s.Namespace.Active)
	assert.Equal(t, []string{"blee", "duh", "crap"}, k9sCfg.K9s.Namespace.Favorites)
}

func TestConfigSave(t *testing.T) {
	path := filepath.Join("/tmp", "k9s.yml")

	k9sCfg.reset()
	k9sCfg.K9s.Namespace.Active = "fred"
	err := k9sCfg.save(path)
	assert.Nil(t, err)

	raw, err := ioutil.ReadFile(path)
	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, string(raw))
}

func TestConfigAddActive(t *testing.T) {
	k9sCfg.reset()
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

	for _, u := range uu {
		k9sCfg.addActive(u.ns)
		assert.Equal(t, u.ns, k9sCfg.K9s.Namespace.Active)
		assert.Equal(t, u.fav, k9sCfg.K9s.Namespace.Favorites)
	}
}

func TestConfigRmFavNS(t *testing.T) {
	uu := []struct {
		ns  string
		fav []string
	}{
		{"all", []string{"default", "kube-system"}},
		{"kube-system", []string{"default"}},
		{"blee", []string{"default"}},
	}

	for _, u := range uu {
		k9sCfg.addActive(u.ns)
		assert.Equal(t, u.ns, k9sCfg.K9s.Namespace.Active)
	}
}

// ----------------------------------------------------------------------------
// Test Data...

var expectedConfig = `k9s:
  refreshRate: 5
  logBufferSize: 200
  namespace:
    active: fred
    favorites:
    - all
    - default
    - kube-system
  view:
    active: po
`
