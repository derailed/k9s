// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNSValidate(t *testing.T) {
	ns := data.NewNamespace()
	ns.Validate(mock.NewMockConnection())

	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"default"}, ns.Recent)
}

func TestNSValidateMissing(t *testing.T) {
	ns := data.NewNamespace()
	ns.Validate(mock.NewMockConnection())

	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"default"}, ns.Recent)
}

func TestNSValidateNoNS(t *testing.T) {
	ns := data.NewNamespace()
	ns.Validate(mock.NewMockConnection())

	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"default"}, ns.Recent)
}

func TestNsValidateMaxNS(t *testing.T) {
	allNS := []string{"ns9", "ns8", "ns7", "ns6", "ns5", "ns4", "ns3", "ns2", "ns1", "all", "default"}
	ns := data.NewNamespace()
	ns.Recent = allNS

	ns.Validate(mock.NewMockConnection())
	assert.Len(t, ns.Recent, data.MaxFavoritesNS)
}

func TestNSAddFav(t *testing.T) {
	ns := data.NewNamespace()

	require.NoError(t, ns.AddFavNS("ns1"))
	require.NoError(t, ns.AddFavNS("default"))

	// Favs own the first slots. Recent ns fills in the leftovers.
	assert.Equal(t, []string{"ns1", "default"}, ns.Favorites)
	assert.Empty(t, ns.Recent)
	assert.Equal(t, []string{"ns1", "default"}, ns.FavNamespaces())
	assert.True(t, ns.IsFav("ns1"))

	ns.RmFavNS("ns1")
	assert.Equal(t, []string{"default"}, ns.Favorites)
}

func TestNSFavSlotsAreStable(t *testing.T) {
	mk := mock.NewMockKubeSettings(makeFlags("cl-1", "ct-1"))
	ns := data.NewNamespace()
	ns.Recent = nil
	require.NoError(t, ns.AddFavNS("fav1"))
	require.NoError(t, ns.AddFavNS("fav2"))

	for i := range data.MaxFavoritesNS {
		require.NoError(t, ns.SetActive(fmt.Sprintf("ns%d", i), mk))
	}

	// Favs kept their slots and the oldest recent ns got recycled.
	assert.Equal(t, []string{"fav1", "fav2"}, ns.Favorites)
	assert.Equal(t,
		[]string{"fav1", "fav2", "ns8", "ns7", "ns6", "ns5", "ns4", "ns3", "ns2"},
		ns.FavNamespaces(),
	)
}

func TestNSMaxFavs(t *testing.T) {
	ns := data.NewNamespace()
	for i := range data.MaxFavoritesNS {
		require.NoError(t, ns.AddFavNS(fmt.Sprintf("ns%d", i)))
	}

	require.Error(t, ns.AddFavNS("blee"))
	// No slots left for recent ns.
	assert.Empty(t, ns.Recent)
	assert.Len(t, ns.FavNamespaces(), data.MaxFavoritesNS)
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

	mk := mock.NewMockKubeSettings(makeFlags("cl-1", "ct-1"))
	ns := data.NewNamespace()
	for _, u := range uu {
		err := ns.SetActive(u.ns, mk)
		require.NoError(t, err)
		assert.Equal(t, u.ns, ns.Active)
		assert.Equal(t, u.fav, ns.Recent)
	}
}

func TestNSValidateRmFavs(t *testing.T) {
	ns := data.NewNamespace()
	ns.Favorites = []string{"default", "fred"}
	ns.Validate(mock.NewMockConnection())

	assert.Equal(t, []string{"default", "fred"}, ns.Favorites)
}
