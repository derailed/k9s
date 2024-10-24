// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port

import (
	"fmt"
	"strings"
)

// PortCheck checks if port is free on host.
type PortChecker func(PortTunnel) bool

// PFAnns represents a collection of port forward annotations.
type PFAnns []*PFAnn

// ToPortSpec returns a container port and local port definitions.
func (aa PFAnns) ToPortSpec(pp ContainerPortSpecs) (string, string) {
	specs, lps := make([]string, 0, len(aa)), make([]string, 0, len(aa))
	for _, a := range aa {
		specs = append(specs, a.AsSpec())
		if a.LocalPort == "" {
			if spec, ok := pp.Find(a); ok {
				a.LocalPort = spec.PortNum
			}
		}
		if a.LocalPort != "" {
			lps = append(lps, a.LocalPort)
		}
	}

	return strings.Join(specs, ","), strings.Join(lps, ",")
}

func (aa PFAnns) ToTunnels(address string, pp ContainerPortSpecs, available PortChecker) (PortTunnels, error) {
	pts := make(PortTunnels, 0, len(aa))
	for _, a := range aa {
		pt, err := a.ToTunnel(address)
		if err != nil {
			return pts, err
		}
		if !available(pt) {
			return pts, fmt.Errorf("Port %s is not available on host", pt.LocalPort)
		}
		pts = append(pts, pt)
	}

	return pts, nil
}

// ParsePFs hydrates a collection of portforward annotations.
func ParsePFs(ann string) (PFAnns, error) {
	ss := strings.Split(ann, ",")
	pp := make(PFAnns, 0, len(ss))
	for _, s := range ss {
		f, err := ParsePF(s)
		if err != nil {
			return nil, err
		}
		pp = append(pp, f)
	}

	return pp, nil
}

func ToTunnels(address, specs, localPorts string) (PortTunnels, error) {
	pp, lps := strings.Split(specs, ","), strings.Split(localPorts, ",")

	if len(pp) != len(lps) {
		return nil, fmt.Errorf("spec to local port count mismatch. Expected %d but got %d", len(pp), len(lps))
	}

	pts := make(PortTunnels, 0, len(pp))
	for i, p := range pp {
		a, err := ParsePF(p)
		if err != nil {
			return nil, err
		}
		n, err := a.PortNum()
		if err != nil {
			return nil, err
		}
		pts = append(pts, PortTunnel{
			Address:       address,
			Container:     a.Container,
			ContainerPort: n,
			LocalPort:     lps[i],
		})
	}

	return pts, nil
}
