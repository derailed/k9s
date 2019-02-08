package config

import (
	"github.com/derailed/k9s/resource"
	log "github.com/sirupsen/logrus"
)

// MaxFavoritesNS number # favorite namespaces to keep in the configuration.
const MaxFavoritesNS = 10

var defaultNamespaces = []string{"all", "default", "kube-system"}

// Namespace tracks active and favorites namespaces.
type Namespace struct {
	Active    string   `yaml:"active"`
	Favorites []string `yaml:"favorites"`
}

// NewNamespace create a new namespace configuration.
func NewNamespace() *Namespace {
	return &Namespace{
		Active:    resource.DefaultNamespace,
		Favorites: defaultNamespaces,
	}
}

// Validate a namespace is setup correctly
func (n *Namespace) Validate(ci ClusterInfo) {
	nn := ci.AllNamespacesOrDie()

	if !n.isAllNamespace() && !InList(nn, n.Active) {
		log.Debugf("[Config] Validation error active namespace reseting to `default")
		n.Active = resource.DefaultNamespace
	}

	for _, ns := range n.Favorites {
		if ns != resource.AllNamespace && !InList(nn, ns) {
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
	return n.Active == resource.AllNamespace
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