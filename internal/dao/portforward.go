package dao

type PortForward struct {
	Generic
}

var _ Accessor = &PortForward{}
var _ Nuker = &PortForward{}

// Delete a portforward.
func (p *PortForward) Delete(path string, cascade, force bool) error {
	p.Factory.DeleteForwarder(path)
	return nil
}
