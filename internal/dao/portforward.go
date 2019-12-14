package dao

import (
	"github.com/rs/zerolog/log"
)

type PortForward struct {
	Generic
}

var _ Accessor = &PortForward{}
var _ Nuker = &PortForward{}

// Delete a portforward.
func (p *PortForward) Delete(path string, cascade, force bool) error {
	log.Debug().Msgf("PortForward DELETE %q", path)
	p.Factory.DeleteForwarder(path)
	return nil
}
