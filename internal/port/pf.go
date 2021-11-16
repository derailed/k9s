package port

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// K9sAutoPortForwardKey represents an auto portforwards annotation.
	K9sAutoPortForwardsKey = "k9scli.io/auto-portforwards"

	// K9sPortForwardKey represents a portforwards annotation.
	K9sPortForwardsKey = "k9scli.io/portforwards"
)

var pfRX = regexp.MustCompile(`\A([\w-]+)::(\d*):?(\d*|[\w-]*)/?(\d+)?\z`)

// PFAnn represents a portforward annotation value.
// Shape: container/portname|portNum:localPort
type PFAnn struct {
	Container        string
	ContainerPort    intstr.IntOrString
	LocalPort        string
	containerPortNum string
}

// ParsePF hydrate a portforward annotation from string.
func ParsePF(ann string) (*PFAnn, error) {
	var pf PFAnn
	r := pfRX.FindStringSubmatch(strings.TrimSpace(ann))
	if len(r) < 4 {
		return &pf, fmt.Errorf("invalid pf annotation %s", ann)
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
