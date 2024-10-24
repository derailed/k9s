// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port

import (
	"fmt"
	"net"
)

// PortTunnels represents a collection of tunnels.
type PortTunnels []PortTunnel

func (t PortTunnels) CheckAvailable() error {
	for _, pt := range t {
		if !IsPortFree(pt) {
			return fmt.Errorf("port %s is not available on host", pt.LocalPort)
		}
	}

	return nil
}

// PortTunnel represents a host tunnel port mapper.
type PortTunnel struct {
	Address, Container, LocalPort, ContainerPort string
}

func NewPortTunnel(a, co, lp, cp string) PortTunnel {
	return PortTunnel{
		Address:       a,
		Container:     co,
		LocalPort:     lp,
		ContainerPort: cp,
	}
}

// String dumps as string.
func (t PortTunnel) String() string {
	return fmt.Sprintf("%s|%s|%s:%s", t.Address, t.Container, t.LocalPort, t.ContainerPort)
}

// PortMap returns a port mapping.
func (t PortTunnel) PortMap() string {
	if t.LocalPort == "" {
		t.LocalPort = t.ContainerPort
	}
	return t.LocalPort + ":" + t.ContainerPort
}

func IsPortFree(t PortTunnel) bool {
	s, err := net.Listen("tcp", fmt.Sprintf("%s:%s", t.Address, t.LocalPort))
	if err != nil {
		return false
	}
	return s.Close() == nil
}
