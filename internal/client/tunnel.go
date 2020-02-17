package client

// PortTunnel represents a host tunnel port mapper.
type PortTunnel struct {
	Address, LocalPort, ContainerPort string
}

// PortMap returns a port mapping.
func (t PortTunnel) PortMap() string {
	return t.LocalPort + ":" + t.ContainerPort
}
