// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/derailed/k9s/internal/slogs"
)

// PortTunnels represents a collection of tunnels.
type PortTunnels []PortTunnel

// CheckAvailable checks if all port tunnels are available.
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

// NewPortTunnel returns a new instance.
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

// IsPortFree checks if a address/port pair is available on host.
func IsPortFree(t PortTunnel) bool {
	s, err := net.Listen("tcp", fmt.Sprintf("%s:%s", t.Address, t.LocalPort))
	if err != nil {
		slog.Warn("Port is not available", slogs.Port, t.LocalPort, slogs.Address, t.Address)
		return false
	}

	return s.Close() == nil
}
