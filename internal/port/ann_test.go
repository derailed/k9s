// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port_test

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/port"
	"github.com/stretchr/testify/assert"
)

func TestPreferredPorts(t *testing.T) {
	uu := map[string]struct {
		anns  port.Annotations
		specs port.ContainerPortSpecs
		err   error
		e     string
	}{
		"no-ports": {
			anns: port.Annotations{
				port.K9sPortForwardsKey: "c1::4321:p1",
			},
			err: errors.New("no exposed ports"),
		},
		"no-annotations": {
			specs: port.ContainerPortSpecs{
				{Container: "c1", PortName: "p1", PortNum: "1234"},
			},
			e: "c1::1234:p1",
		},
		"single-numb": {
			anns: port.Annotations{
				port.K9sPortForwardsKey: "c1::4321:1234",
			},
			specs: port.ContainerPortSpecs{
				{Container: "c1", PortName: "p1", PortNum: "1234"},
			},
			e: "c1::4321:1234/1234",
		},
		"single-same": {
			anns: port.Annotations{
				port.K9sPortForwardsKey: "c1::1234",
			},
			specs: port.ContainerPortSpecs{
				{Container: "c1", PortName: "p1", PortNum: "1234"},
			},
			e: "c1::1234:1234/1234",
		},
		"single-mismatch": {
			anns: port.Annotations{
				port.K9sPortForwardsKey: "c2::4321:p1",
			},
			specs: port.ContainerPortSpecs{
				{Container: "c1", PortName: "p1", PortNum: "1234"},
			},
		},
		"multi": {
			anns: port.Annotations{
				port.K9sPortForwardsKey: "c1::4321:1234,c1::5432:2345",
			},
			specs: port.ContainerPortSpecs{
				{Container: "c1", PortName: "p1", PortNum: "1234"},
				{Container: "c1", PortName: "p2", PortNum: "2345"},
			},
			e: "c1::4321:1234/1234,c1::5432:2345/2345",
		},
		"multi-mismatch": {
			anns: port.Annotations{
				port.K9sPortForwardsKey: "c1::4321:1234,c1::5432:2345",
			},
			specs: port.ContainerPortSpecs{
				{Container: "c1", PortName: "p1", PortNum: "1234"},
				{Container: "c2", PortName: "p3", PortNum: "2345"},
			},
			e: "c1::4321:1234/1234",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			anns, err := u.anns.PreferredPorts(u.specs)
			assert.Equal(t, u.err, err)
			if err != nil {
				return
			}
			pfs, err := port.ParsePFs(u.e)
			if err != nil {
				pfs = port.PFAnns{}
			}
			assert.Equal(t, pfs, anns)
		})
	}
}
