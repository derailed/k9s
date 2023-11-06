package watch

import (
	"strings"

	"github.com/derailed/k9s/internal/port"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/portforward"
)

// Forwarder represents a port forwarder.
type Forwarder interface {
	// Start starts a port-forward.
	Start(path string, tunnel port.PortTunnel) (*portforward.PortForwarder, error)

	// Stop terminates a port forward.
	Stop()

	// ID returns the pf id.
	ID() string

	// Container returns a container name.
	Container() string

	// Ports returns the port mapping.
	Port() string

	// FQN returns the full port-forward name.
	FQN() string

	// Active returns forwarder current state.
	Active() bool

	// SetActive sets port-forward state.
	SetActive(bool)

	// Age returns forwarder age.
	Age() string

	// HasPortMapping returns true if port mapping exists.
	HasPortMapping(string) bool
}

// Forwarders tracks active port forwards.
type Forwarders map[string]Forwarder

// NewForwarders returns new forwarders.
func NewForwarders() Forwarders {
	return make(map[string]Forwarder)
}

// BOZO!! Review!!!
// IsPodForwarded checks if pod has a forward.
func (ff Forwarders) IsPodForwarded(fqn string) bool {
	for k := range ff {
		if strings.HasPrefix(k, fqn) {
			return true
		}
	}

	return false
}

// IsContainerForwarded checks if pod has a forward.
func (ff Forwarders) IsContainerForwarded(fqn, co string) bool {
	prefix := fqn + "|" + co
	for k := range ff {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}

	return false
}

// DeleteAll stops and delete all port-forwards.
func (ff Forwarders) DeleteAll() {
	for k, f := range ff {
		log.Debug().Msgf("Deleting forwarder %s", f.ID())
		f.Stop()
		delete(ff, k)
	}
}

// Kill stops and delete a port-forwards associated with pod.
func (ff Forwarders) Kill(path string) int {
	var stats int
	for k, f := range ff {
		if strings.HasPrefix(k, path) {
			stats++
			log.Debug().Msgf("Stop + Delete port-forward %s", k)
			f.Stop()
			delete(ff, k)
		}
	}

	return stats
}

// Dump for debug!
func (ff Forwarders) Dump() {
	log.Debug().Msgf("----------- PORT-FORWARDS --------------")
	for k, f := range ff {
		log.Debug().Msgf("  %s -- %#v", k, f)
	}
}
