// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
)

const (
	// MaxFavoritesNS number # favorite namespaces to keep in the configuration.
	MaxFavoritesNS = 9
)

// Namespace tracks active, favorites and recently used namespaces.
// Favorites are user managed and own the first slots. Recent ones fill in the
// leftovers and get recycled oldest first.
type Namespace struct {
	Active string `yaml:"active"`
	// Deprecated: only read to migrate legacy configurations.
	LockFavorites *bool    `yaml:"lockFavorites,omitempty"`
	Favorites     []string `yaml:"favorites"`
	Recent        []string `yaml:"recent"`
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
		Active: n,
		Recent: []string{client.DefaultNamespace},
	}
}

// migrate tracks legacy auto-assigned favorites as recent namespaces.
func (n *Namespace) migrate() {
	if n.LockFavorites == nil {
		return
	}
	if !*n.LockFavorites {
		n.Favorites, n.Recent = nil, n.Favorites
	}
	n.LockFavorites = nil
}

func (n *Namespace) merge(old *Namespace) {
	n.mx.Lock()
	defer n.mx.Unlock()

	for _, fav := range old.Favorites {
		if !slices.Contains(n.Favorites, fav) {
			n.Favorites = append(n.Favorites, fav)
		}
	}
	for _, ns := range old.Recent {
		if !slices.Contains(n.Favorites, ns) && !slices.Contains(n.Recent, ns) {
			n.Recent = append(n.Recent, ns)
		}
	}

	n.trimFavNs()
}

// Validate validates a namespace is setup correctly.
func (n *Namespace) Validate(conn client.Connection) {
	n.mx.Lock()
	defer n.mx.Unlock()

	if conn == nil || !conn.IsValidNamespace(n.Active) {
		return
	}
	for _, ns := range slices.Concat(n.Favorites, n.Recent) {
		if !conn.IsValidNamespace(ns) {
			slog.Debug("Invalid favorite found",
				slogs.Namespace, ns,
				slogs.AllNS, n.isAllNamespaces(),
			)
			n.rmFavNS(ns)
		}
	}

	n.trimFavNs()
}

// SetActive set the active namespace.
func (n *Namespace) SetActive(ns string, _ KubeSettings) error {
	if n == nil {
		n = NewActiveNamespace(ns)
	}

	n.mx.Lock()
	defer n.mx.Unlock()

	if ns == client.BlankNamespace {
		ns = client.NamespaceAll
	}
	n.Active = ns

	if ns != "" {
		n.addRecentNS(ns)
	}

	return nil
}

func (n *Namespace) isAllNamespaces() bool {
	return n.Active == client.NamespaceAll || n.Active == ""
}

// FavNamespaces returns the favorite namespaces in slot order.
func (n *Namespace) FavNamespaces() []string {
	n.mx.RLock()
	defer n.mx.RUnlock()

	return slices.Concat(n.Favorites, n.Recent)
}

// AddFavNS adds a namespace to favorites if not already present.
func (n *Namespace) AddFavNS(ns string) error {
	n.mx.Lock()
	defer n.mx.Unlock()

	if slices.Contains(n.Favorites, ns) {
		return nil
	}
	if len(n.Favorites) >= MaxFavoritesNS {
		return fmt.Errorf("unable to fav %q. Only %d favorites are allowed", ns, MaxFavoritesNS)
	}
	n.Favorites = append(n.Favorites, ns)
	n.rmRecentNS(ns)
	n.trimFavNs()

	return nil
}

func (n *Namespace) addRecentNS(ns string) {
	if slices.Contains(n.Favorites, ns) || slices.Contains(n.Recent, ns) {
		return
	}

	n.Recent = append([]string{ns}, n.Recent...)
	n.trimFavNs()
}

// RmFavNS removes a namespace from favorites.
func (n *Namespace) RmFavNS(ns string) {
	n.mx.Lock()
	defer n.mx.Unlock()

	n.rmFavNS(ns)
}

func (n *Namespace) rmFavNS(ns string) {
	if i := slices.Index(n.Favorites, ns); i >= 0 {
		n.Favorites = slices.Delete(n.Favorites, i, i+1)
	}
	n.rmRecentNS(ns)
}

func (n *Namespace) rmRecentNS(ns string) {
	if i := slices.Index(n.Recent, ns); i >= 0 {
		n.Recent = slices.Delete(n.Recent, i, i+1)
	}
}

// IsFav checks if a namespace is in the favorites list.
func (n *Namespace) IsFav(ns string) bool {
	n.mx.RLock()
	defer n.mx.RUnlock()

	return slices.Contains(n.Favorites, ns)
}

// trimFavNs caps favorites and recycles the oldest recent namespaces to fit in
// the leftover slots.
func (n *Namespace) trimFavNs() {
	if len(n.Favorites) > MaxFavoritesNS {
		slog.Debug("Number of favorite exceeds hard limit. Trimming.", slogs.Max, MaxFavoritesNS)
		n.Favorites = n.Favorites[:MaxFavoritesNS]
	}
	if free := MaxFavoritesNS - len(n.Favorites); len(n.Recent) > free {
		n.Recent = n.Recent[:free]
	}
}
