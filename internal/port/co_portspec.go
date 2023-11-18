// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port

import (
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ContainerPortSpecs represents a container exposed ports.
type ContainerPortSpecs []ContainerPortSpec

func (c ContainerPortSpecs) Dump() string {
	ss := make([]string, 0, len(c))
	for _, spec := range c {
		ss = append(ss, spec.String())
	}

	return strings.Join(ss, "\n")
}

// InSpecs checks if given port matches a spec.
func (c ContainerPortSpecs) MatchSpec(s string) bool {
	// Skip validation if No port are exposed or no container port spec.
	if len(c) == 0 || !strings.Contains(s, "::") {
		return true
	}
	for _, spec := range c {
		if spec.MatchSpec(s) {
			return true
		}
	}

	return false
}

// ToTunnels convert port specs to tunnels.
func (c ContainerPortSpecs) ToTunnels(address string) PortTunnels {
	tt := make(PortTunnels, 0, len(c))
	for _, spec := range c {
		tt = append(tt, spec.ToTunnel(address))
	}

	return tt
}

// Find finds a matching container port.
func (c ContainerPortSpecs) Find(pf *PFAnn) (ContainerPortSpec, bool) {
	for _, spec := range c {
		if spec.Match(pf) {
			return spec, true
		}
	}

	return ContainerPortSpec{}, false
}

// Match checks if container ports match a pf annotation.
func (c ContainerPortSpecs) Match(pf *PFAnn) bool {
	for _, spec := range c {
		if spec.Match(pf) {
			return true
		}
	}

	return false
}

func (c ContainerPortSpecs) MatchAnnotations(s string) PFAnns {
	pfs, err := ParsePFs(s)
	if err != nil {
		return nil
	}

	mm := make(PFAnns, 0, len(c))
	for _, pf := range pfs {
		if pf.Match(c) {
			mm = append(mm, pf)
		}
	}

	return mm
}

// FromContainerPorts hydrates from a pod container specification.
func FromContainerPorts(co string, pp []v1.ContainerPort) ContainerPortSpecs {
	specs := make(ContainerPortSpecs, 0, len(pp))
	for _, p := range pp {
		if p.Protocol != v1.ProtocolTCP {
			continue
		}
		specs = append(specs, NewPortSpec(co, p.Name, p.ContainerPort))
	}

	return specs
}

// ContainerPortSpec represents a container port specification.
type ContainerPortSpec struct {
	Container string
	PortName  string
	PortNum   string
}

// NewPortSpec returns a new instance.
func NewPortSpec(co, portName string, port int32) ContainerPortSpec {
	return ContainerPortSpec{
		Container: co,
		PortName:  portName,
		PortNum:   strconv.Itoa(int(port)),
	}
}

func (c ContainerPortSpec) MatchSpec(s string) bool {
	tokens := strings.Split(s, "::")
	if len(tokens) < 2 {
		return false
	}

	return tokens[0] == c.Container && tokens[1] == c.PortNum
}

func (c ContainerPortSpec) ToTunnel(address string) PortTunnel {
	return PortTunnel{
		Address:       address,
		LocalPort:     c.PortNum,
		ContainerPort: c.PortNum,
	}
}

func (c ContainerPortSpec) Port() intstr.IntOrString {
	if c.PortName != "" {
		return intstr.Parse(c.PortName)
	}

	return intstr.Parse(c.PortNum)
}

func (c ContainerPortSpec) ToPFAnn() *PFAnn {
	return &PFAnn{
		Container:     c.Container,
		ContainerPort: c.Port(),
		LocalPort:     c.PortNum,
	}
}

// Match checks if the container spec matches an annotation.
func (c ContainerPortSpec) Match(ann *PFAnn) bool {
	if c.Container != ann.Container {
		return false
	}

	switch ann.ContainerPort.Type {
	case intstr.String:
		return c.PortName == ann.ContainerPort.String()
	case intstr.Int:
		return c.PortNum == ann.ContainerPort.String()
	default:
		return false
	}
}

// String dumps spec to string.
func (c ContainerPortSpec) String() string {
	s := c.Container + "::" + c.PortNum
	if c.PortName != "" {
		s += "(" + c.PortName + ")"
	}
	return s
}
