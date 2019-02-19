package config

import (
	log "github.com/sirupsen/logrus"
)

const (
	// MaxFavoritesNS number # favorite namespaces to keep in the configuration.
	MaxFavoritesNS = 10
	defaultNS      = "default"
	allNS          = "all"
)

// Namespace tracks active and favorites namespaces.
type Namespace struct {
	Active    string   `yaml:"active"`
	Favorites []string `yaml:"favorites"`
}

// NewNamespace create a new namespace configuration.
func NewNamespace() *Namespace {
	return &Namespace{
		Active:    defaultNS,
		Favorites: []string{defaultNS},
	}
}

// Validate a namespace is setup correctly
func (n *Namespace) Validate(ks KubeSettings) {
	nn, err := ks.NamespaceNames()
	if err != nil {
		return
	}

	if !n.isAllNamespace() && !InList(nn, n.Active) {
		log.Debugf("[Config] Validation error active namespace resetting to `default")
		n.Active = defaultNS
	}

	for _, ns := range n.Favorites {
		if ns != allNS && !InList(nn, ns) {
			log.Debugf("[Config] Invalid favorite found '%s' - %t", ns, n.isAllNamespace())
			n.rmFavNS(ns)
		}
	}
}

// SetActive set the active namespace.
func (n *Namespace) SetActive(ns string) {
	n.Active = ns
	n.addFavNS(ns)
}

func (n *Namespace) isAllNamespace() bool {
	return n.Active == allNS
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
