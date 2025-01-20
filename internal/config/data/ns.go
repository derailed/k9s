// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
)

const (
	// MaxFavoritesNS number # favorite namespaces to keep in the configuration.
	MaxFavoritesNS = 9
)

// Namespace tracks active and favorites namespaces.
type Namespace struct {
	Active        string   `yaml:"active"`
	LockFavorites bool     `yaml:"lockFavorites"`
	Favorites     []string `yaml:"favorites"`
	mx            sync.RWMutex
}

// NewNamespace create a new namespace configuration.
func NewNamespace() *Namespace {
	return NewActiveNamespace(client.DefaultNamespace)
}

func NewActiveNamespace(n string) *Namespace {
	if n == client.BlankNamespace {
		n = client.DefaultNamespace
	}

	return &Namespace{
		Active:    n,
		Favorites: []string{client.DefaultNamespace},
	}
}

func (n *Namespace) merge(old *Namespace) {
	n.mx.Lock()
	defer n.mx.Unlock()

	if n.LockFavorites {
		return
	}
	for _, fav := range old.Favorites {
		if InList(n.Favorites, fav) {
			continue
		}
		n.Favorites = append(n.Favorites, fav)
	}

	n.trimFavNs()
}

// Validate validates a namespace is setup correctly.
func (n *Namespace) Validate(c client.Connection) {
	n.mx.RLock()
	defer n.mx.RUnlock()

	if c == nil || !c.IsValidNamespace(n.Active) {
		return
	}
	for _, ns := range n.Favorites {
		if !c.IsValidNamespace(ns) {
			log.Debug().Msgf("[Namespace] Invalid favorite found '%s' - %t", ns, n.isAllNamespaces())
			n.rmFavNS(ns)
		}
	}

	n.trimFavNs()
}

// SetActive set the active namespace.
func (n *Namespace) SetActive(ns string, ks KubeSettings) error {
	if n == nil {
		n = NewActiveNamespace(ns)
	}

	n.mx.Lock()
	defer n.mx.Unlock()

	if ns == client.BlankNamespace {
		ns = client.NamespaceAll
	}
	n.Active = ns

	if ns != "" && !n.LockFavorites {
		n.addFavNS(ns)
	}

	return nil
}

func (n *Namespace) isAllNamespaces() bool {
	return n.Active == client.NamespaceAll || n.Active == ""
}

func (n *Namespace) addFavNS(ns string) {
	if InList(n.Favorites, ns) {
		return
	}

	nfv := make([]string, 0, MaxFavoritesNS)
	nfv = append(nfv, ns)
	for i := 0; i < len(n.Favorites); i++ {
		if i+1 < MaxFavoritesNS {
			nfv = append(nfv, n.Favorites[i])
		}
	}
	n.Favorites = nfv
}

func (n *Namespace) rmFavNS(ns string) {
	if n.LockFavorites {
		return
	}

	victim := -1
	for i, f := range n.Favorites {
		if f == ns {
			victim = i
			break
		}
	}
	if victim < 0 {
		return
	}

	n.Favorites = append(n.Favorites[:victim], n.Favorites[victim+1:]...)
}

func (n *Namespace) trimFavNs() {
	if len(n.Favorites) > MaxFavoritesNS {
		log.Debug().Msgf("[Namespace] Number of favorite exceeds hard limit of %v. Trimming.", MaxFavoritesNS)
		n.Favorites = n.Favorites[:MaxFavoritesNS]
	}
}
