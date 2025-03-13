// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/stretchr/testify/assert"
)

func TestNSValidate(t *testing.T) {
	ns := data.NewNamespace()
	ns.Validate(mock.NewMockConnection())

	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"default"}, ns.Favorites)
}

func TestNSValidateMissing(t *testing.T) {
	ns := data.NewNamespace()
	ns.Validate(mock.NewMockConnection())

	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"default"}, ns.Favorites)
}

func TestNSValidateNoNS(t *testing.T) {
	ns := data.NewNamespace()
	ns.Validate(mock.NewMockConnection())

	assert.Equal(t, "default", ns.Active)
	assert.Equal(t, []string{"default"}, ns.Favorites)
}

func TestNsValidateMaxNS(t *testing.T) {
	allNS := []string{"ns9", "ns8", "ns7", "ns6", "ns5", "ns4", "ns3", "ns2", "ns1", "all", "default"}
	ns := data.NewNamespace()
	ns.Favorites = allNS

	ns.Validate(mock.NewMockConnection())
	assert.Equal(t, data.MaxFavoritesNS, len(ns.Favorites))
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
		assert.Nil(t, err)
		assert.Equal(t, u.ns, ns.Active)
		assert.Equal(t, u.fav, ns.Favorites)
	}
}

func TestNSValidateRmFavs(t *testing.T) {
	ns := data.NewNamespace()
	ns.Favorites = []string{"default", "fred"}
	ns.Validate(mock.NewMockConnection())

	assert.Equal(t, []string{"default", "fred"}, ns.Favorites)
}
