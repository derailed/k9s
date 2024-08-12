// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port_test

import (
	"testing"

	"github.com/derailed/k9s/internal/port"
	"github.com/stretchr/testify/assert"
)

func TestContainerPortSpecMatch(t *testing.T) {
	uu := map[string]struct {
		ann  string
		spec port.ContainerPortSpec
		e    bool
	}{
		"full": {
			ann: "c1::4321:1234",
			spec: port.ContainerPortSpec{
				Container: "c1",
				PortNum:   "1234",
			},
			e: true,
		},
		"no-port-name": {
			ann: "c1::4321:p1/1234",
			spec: port.ContainerPortSpec{
				Container: "c1",
				PortName:  "p1",
				PortNum:   "1234",
			},
			e: true,
		},
		"port-name-hosed": {
			ann: "c1::4321:blee/1234",
			spec: port.ContainerPortSpec{
				Container: "c1",
				PortName:  "fred",
				PortNum:   "1234",
			},
		},
		"container-name-hosed": {
			ann: "c2::4321:fred/1234",
			spec: port.ContainerPortSpec{
				Container: "c1",
				PortName:  "blee",
				PortNum:   "1234",
			},
		},
		"port-num-hosed": {
			ann: "c2::4321:1235",
			spec: port.ContainerPortSpec{
				Container: "c1",
				PortNum:   "1234",
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pf, err := port.ParsePF(u.ann)
			assert.Nil(t, err)

			assert.Equal(t, u.e, u.spec.Match(pf))
		})
	}
}

func TestContainerPortSpecString(t *testing.T) {
	uu := map[string]struct {
		spec port.ContainerPortSpec
		e    string
	}{
		"full": {
			spec: port.NewPortSpec("c1", "p1", 1234),
			e:    "c1::1234(p1)",
		},
		"no-name": {
			spec: port.NewPortSpec("c1", "", 1234),
			e:    "c1::1234",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.spec.String())
		})
	}
}

func TestContainerPortSpecsMatch(t *testing.T) {
	uu := map[string]struct {
		ann   string
		specs port.ContainerPortSpecs
		e     bool
	}{
		"full": {
			ann: "c1::4321:p1",
			specs: port.ContainerPortSpecs{
				port.NewPortSpec("c1", "p1", 1234),
				port.NewPortSpec("c2", "p2", 1235),
			},
			e: true,
		},
		"no-name": {
			ann: "c1::4321",
			specs: port.ContainerPortSpecs{
				port.NewPortSpec("c1", "", 4321),
				port.NewPortSpec("c2", "p2", 1235),
			},
			e: true,
		},
		"name-hosed": {
			ann: "c1::4321:p4",
			specs: port.ContainerPortSpecs{
				port.NewPortSpec("c1", "p1", 1234),
				port.NewPortSpec("c2", "p2", 1235),
			},
		},
		"numb-hosed": {
			ann: "c1::4321:1235",
			specs: port.ContainerPortSpecs{
				port.NewPortSpec("c1", "p1", 1234),
				port.NewPortSpec("c2", "p2", 1236),
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pf, err := port.ParsePF(u.ann)
			assert.Nil(t, err)
			assert.Equal(t, u.e, u.specs.Match(pf))
		})
	}
}
