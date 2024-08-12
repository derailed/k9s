// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port_test

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/port"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestParsePFs(t *testing.T) {
	uu := map[string]struct {
		spec string
		pfs  port.PFAnns
		e    error
	}{
		"single": {
			spec: "c2::4321:1234",
			pfs: port.PFAnns{
				{Container: "c2", ContainerPort: intstr.Parse("1234"), LocalPort: "4321"},
			},
		},
		"multi": {
			spec: "c1::4321:1234,c2::6666:6543",
			pfs: port.PFAnns{
				{Container: "c1", ContainerPort: intstr.Parse("1234"), LocalPort: "4321"},
				{Container: "c2", ContainerPort: intstr.Parse("6543"), LocalPort: "6666"},
			},
		},
		"spaces": {
			spec: " c1::4321:1234 , c2::6666:6543 ",
			pfs: port.PFAnns{
				{Container: "c1", ContainerPort: intstr.Parse("1234"), LocalPort: "4321"},
				{Container: "c2", ContainerPort: intstr.Parse("6543"), LocalPort: "6666"},
			},
		},
		"plain-multi": {
			spec: "4321:1234, 6666:6543",
			pfs: port.PFAnns{
				{ContainerPort: intstr.Parse("1234"), LocalPort: "4321"},
				{ContainerPort: intstr.Parse("6543"), LocalPort: "6666"},
			},
		},
		"toast": {
			spec: "c1::p1:1234,c2::4321",
			e:    errors.New("invalid port-forward specification c1::p1:1234"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pfs, err := port.ParsePFs(u.spec)
			assert.Equal(t, u.e, err)
			if err != nil {
				return
			}
			assert.Equal(t, u.pfs, pfs)
		})
	}
}

func TestPFsToTunnel(t *testing.T) {
	uu := map[string]struct {
		exp   string
		specs port.ContainerPortSpecs
		pts   port.PortTunnels
		e     error
	}{
		"single": {
			exp: "c2::4321:1234",
			specs: port.ContainerPortSpecs{
				{Container: "c2", PortName: "p1", PortNum: "1234"},
			},
			pts: port.PortTunnels{
				{Address: "fred", Container: "c2", ContainerPort: "1234", LocalPort: "4321"},
			},
		},
		"hosed": {
			exp: "c2::p2",
			specs: port.ContainerPortSpecs{
				{Container: "c2", PortName: "p1", PortNum: "1234"},
			},
			pts: port.PortTunnels{
				{Address: "fred", Container: "c2", ContainerPort: "1234", LocalPort: "4321"},
			},
			e: errors.New("no port number assigned"),
		},
	}

	f := func(port.PortTunnel) bool {
		return true
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pfs, err := port.ParsePFs(u.exp)
			assert.Nil(t, err)
			pts, err := pfs.ToTunnels("fred", u.specs, f)
			assert.Equal(t, u.e, err)
			if err != nil {
				return
			}
			assert.Equal(t, u.pts, pts)
		})
	}
}

func TestPFsToPortSpec(t *testing.T) {
	uu := map[string]struct {
		exp        string
		spec, port string
		specs      port.ContainerPortSpecs
		e          error
	}{
		"single": {
			exp:  "c2::4321:p2/1234",
			spec: "c2::1234",
			port: "4321",
			specs: port.ContainerPortSpecs{
				{Container: "c2", PortNum: "1234"},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pfs, err := port.ParsePFs(u.exp)
			assert.Equal(t, u.e, err)
			if err != nil {
				return
			}
			spec, port := pfs.ToPortSpec(u.specs)
			assert.Equal(t, u.spec, spec)
			assert.Equal(t, u.port, port)
		})
	}
}

func TestToTunnels(t *testing.T) {
	uu := map[string]struct {
		specs, ports string
		tunnels      port.PortTunnels
		err          error
	}{
		"single": {
			specs: "c2::4321:p2/1234",
			ports: "4321",
			tunnels: port.PortTunnels{
				{
					Address:       "blee",
					LocalPort:     "4321",
					Container:     "c2",
					ContainerPort: "1234",
				},
			},
		},
		"multi": {
			specs: "c1::5432:2345/2345,c2::4321:p2/1234",
			ports: "5432,4321",
			tunnels: port.PortTunnels{
				{
					Address:       "blee",
					LocalPort:     "5432",
					Container:     "c1",
					ContainerPort: "2345",
				},
				{
					Address:       "blee",
					LocalPort:     "4321",
					Container:     "c2",
					ContainerPort: "1234",
				},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			tt, err := port.ToTunnels("blee", u.specs, u.ports)
			assert.Equal(t, u.err, err)
			if err != nil {
				return
			}
			assert.Equal(t, u.tunnels, tt)
		})
	}
}
