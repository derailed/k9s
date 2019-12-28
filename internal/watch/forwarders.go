package watch

import (
	"strings"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/portforward"
)

// Forwarder represents a port forwarder.
type Forwarder interface {
	// Start initializes a port forward.
	Start(path, co string, ports []string) (*portforward.PortForwarder, error)

	// Stop terminates a port forward.
	Stop()

	// Path returns a resource FQN.
	Path() string

	// Container returns a container name.
	Container() string

	// Ports returns container exposed ports.
	Ports() []string

	// Active returns forwarder current state.
	Active() bool

	// Age returns forwarder age.
	Age() string
}

// Forwarders tracks active port forwards.
type Forwarders map[string]Forwarder

// NewForwarders returns new forwarders.
func NewForwarders() Forwarders {
	return make(map[string]Forwarder)
}

// KillAll stops and delete all port-forwards.
func (ff Forwarders) DeleteAll() {
	ff.Dump()
	for k, f := range ff {
		log.Debug().Msgf("Deleting forwarder %s", f.Path())
		f.Stop()
		delete(ff, k)
	}
}

// Kill stops and delete a port-forwards associated with pod.
func (ff Forwarders) Kill(path string) int {
	ff.Dump()

	log.Debug().Msgf("Delete port-forward %q", path)
	hasContainer := strings.Contains(path, ":")
	var stats int
	for k, f := range ff {
		victim := k
		if !hasContainer {
			victim = strings.Split(k, ":")[0]
		}
		if victim == path {
			stats++
			log.Debug().Msgf("Stopping and delete port-forward %s", k)
			f.Stop()
			delete(ff, k)
		}
	}

	return stats
}

func (ff Forwarders) Dump() {
	log.Debug().Msgf("----------- PORT-FORWARDS --------------")
	for k, f := range ff {
		log.Debug().Msgf("  %s -- %#v", k, f)
	}
}
