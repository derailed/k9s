// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// K9sAutoPortForwardsKey represents an auto portforwards annotation.
	K9sAutoPortForwardsKey = "k9scli.io/auto-port-forwards"

	// K9sPortForwardsKey represents a portforwards annotation.
	K9sPortForwardsKey = "k9scli.io/port-forwards"
)

var (
	pfRX      = regexp.MustCompile(`\A([\w-]+)::(\d*):?(\d*|[\w-]*)/?(\d+)?\z`)
	pfPlainRX = regexp.MustCompile(`\A(\d*):?(\d*|[\w-]*)\z`)
)

// PFAnn represents a portforward annotation value.
// Shape: container/portname|portNum:localPort
type PFAnn struct {
	Container        string
	ContainerPort    intstr.IntOrString
	LocalPort        string
	containerPortNum string
}

func ParsePlainPF(ann string) (*PFAnn, error) {
	if len(ann) == 0 {
		return nil, fmt.Errorf("invalid annotation %q", ann)
	}
	var pf PFAnn
	mm := pfPlainRX.FindStringSubmatch(strings.TrimSpace(ann))
	if len(mm) < 3 {
		return nil, fmt.Errorf("Invalid plain port-forward %s", ann)
	}
	if len(mm[2]) == 0 {
		pf.ContainerPort = intstr.Parse(mm[1])
		pf.LocalPort = mm[1]
		return &pf, nil
	}
	pf.LocalPort, pf.ContainerPort = mm[1], intstr.Parse(mm[2])

	return &pf, nil
}

// ParsePF hydrate a portforward annotation from string.
func ParsePF(ann string) (*PFAnn, error) {
	if pf, err := ParsePlainPF(ann); err == nil {
		return pf, nil
	}
	var pf PFAnn
	if mm := pfPlainRX.FindStringSubmatch(strings.TrimSpace(ann)); len(mm) == 3 {
		pf.containerPortNum = mm[0]
	}
	r := pfRX.FindStringSubmatch(strings.TrimSpace(ann))
	if len(r) < 4 {
		return &pf, fmt.Errorf("invalid port-forward specification %s", ann)
	}
	pf.Container = r[1]
	pf.LocalPort, pf.ContainerPort = r[2], intstr.Parse(r[3])
	if r[3] == "" {
		pf.ContainerPort = intstr.Parse(pf.LocalPort)
	}

	// Testing only!
	if len(r) == 5 && r[4] != "" {
		pf.containerPortNum = r[4]
	}
	if pf.LocalPort == "" {
		pf.LocalPort = pf.containerPortNum
	}

	return &pf, nil
}

// Match checks if annotation matches any of the container ports.
func (p *PFAnn) Match(ss ContainerPortSpecs) bool {
	for _, s := range ss {
		if s.Match(p) {
			p.containerPortNum = s.PortNum
			return true
		}
	}

	return false
}

func (p *PFAnn) AsSpec() string {
	s := p.Container + "::"
	if p.containerPortNum != "" {
		return s + p.containerPortNum
	}
	return s + p.LocalPort
}

// String dumps the annotation.
func (p *PFAnn) String() string {
	return p.Container + "::" + p.LocalPort + ":" + p.containerPortNum
}

func (p *PFAnn) PortNum() (string, error) {
	if p.ContainerPort.Type == intstr.Int {
		return p.ContainerPort.String(), nil
	}
	if p.containerPortNum != "" {
		return p.containerPortNum, nil
	}

	return "", errors.New("no port number assigned")
}

func (p *PFAnn) ToTunnel(address string) (PortTunnel, error) {
	var pt PortTunnel
	port, err := p.PortNum()
	if err != nil {
		return pt, err
	}

	pt.Address, pt.Container = address, p.Container
	pt.ContainerPort, pt.LocalPort = port, p.LocalPort

	return pt, nil
}
