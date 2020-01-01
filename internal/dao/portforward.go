package dao

import (
	"github.com/derailed/k9s/internal/client"
)

// PortForward represents a port forward dao.
type PortForward struct {
	Generic
}

var _ Accessor = &PortForward{}
var _ Nuker = &PortForward{}

// Delete a portforward.
func (p *PortForward) Delete(path string, cascade, force bool) error {
	ns, _ := client.Namespaced(path)
	auth, err := p.Client().CanI(ns, "v1/pods:portforward", []string{"delete"})
	if !auth || err != nil {
		return err
	}
	p.Factory.DeleteForwarder(path)

	return nil
}
