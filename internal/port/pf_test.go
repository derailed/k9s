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

func TestParsePF(t *testing.T) {
	uu := map[string]struct {
		exp           string
		container     string
		containerPort intstr.IntOrString
		localPort     string
		e             error
	}{
		"full-numbs": {
			exp:           "c1::4321:1234",
			container:     "c1",
			containerPort: intstr.Parse("1234"),
			localPort:     "4321",
		},
		"full-named": {
			exp:           "c1::4321:p1/1234",
			container:     "c1",
			containerPort: intstr.Parse("p1"),
			localPort:     "4321",
		},
		"just-named": {
			exp:           "c1::p1/1234",
			container:     "c1",
			containerPort: intstr.Parse("p1"),
			localPort:     "1234",
		},
		"just-num": {
			exp:           "c1::1234",
			container:     "c1",
			containerPort: intstr.Parse("1234"),
			localPort:     "1234",
		},
		"plain-single": {
			exp:           "1234",
			container:     "",
			containerPort: intstr.Parse("1234"),
			localPort:     "1234",
		},
		"plain-full": {
			exp:           "4321:1234",
			container:     "",
			containerPort: intstr.Parse("1234"),
			localPort:     "4321",
		},
		"toast": {
			exp: "c1:4321:1234",
			e:   errors.New("invalid port-forward specification c1:4321:1234"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pf, err := port.ParsePF(u.exp)
			assert.Equal(t, u.e, err)
			if err != nil {
				return
			}
			assert.Equal(t, u.container, pf.Container)
			assert.Equal(t, u.containerPort, pf.ContainerPort)
			assert.Equal(t, u.localPort, pf.LocalPort)
		})
	}
}

func TestPFMatch(t *testing.T) {
	uu := map[string]struct {
		exp   string
		specs port.ContainerPortSpecs
		err   error
		e     bool
	}{
		"match": {
			exp: "c1::1234",
			specs: port.ContainerPortSpecs{
				{Container: "c1", PortNum: "1234"},
			},
			e: true,
		},
		"match-portnum": {
			exp: "c1::4321:1234",
			specs: port.ContainerPortSpecs{
				{Container: "c1", PortNum: "1234"},
			},
			e: true,
		},
		"no-match": {
			exp: "c1::1235",
			specs: port.ContainerPortSpecs{
				{Container: "c1", PortNum: "1234"},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pf, err := port.ParsePF(u.exp)
			assert.Equal(t, u.err, err)
			if err != nil {
				return
			}
			assert.Equal(t, u.e, pf.Match(u.specs))
		})
	}
}

func TestPFPortNum(t *testing.T) {
	uu := map[string]struct {
		exp string
		err error
		e   string
	}{
		"port-name": {
			exp: "c1::4321:1234",
			e:   "1234",
		},
		"port-number": {
			exp: "c1::4321:1234",
			e:   "1234",
		},
		"missing-port-number": {
			exp: "c1::p1",
			err: errors.New("no port number assigned"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pf, err := port.ParsePF(u.exp)
			assert.Nil(t, err)
			n, err := pf.PortNum()
			assert.Equal(t, u.err, err)
			if err != nil {
				return
			}
			assert.Equal(t, u.e, n)
		})
	}
}

func TestPFToTunnel(t *testing.T) {
	uu := map[string]struct {
		exp string
		err error
		e   port.PortTunnel
	}{
		"port-name": {
			exp: "c1::p1/1234",
			e: port.PortTunnel{
				Address:       "blee",
				Container:     "c1",
				LocalPort:     "1234",
				ContainerPort: "1234",
			},
		},
		"port-numb": {
			exp: "c1::4321:1234",
			e: port.PortTunnel{
				Address:       "blee",
				Container:     "c1",
				LocalPort:     "4321",
				ContainerPort: "1234",
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pf, err := port.ParsePF(u.exp)
			assert.Nil(t, err)
			pt, err := pf.ToTunnel("blee")
			assert.Equal(t, u.err, err)
			if err != nil {
				return
			}
			assert.Equal(t, u.e, pt)
		})
	}
}

func TestPFString(t *testing.T) {
	uu := map[string]struct {
		exp string
		err error
		e   string
	}{
		"port-name": {
			exp: "c1::p1/1234",
			e:   "c1::1234:1234",
		},
		"port-numb": {
			exp: "c1::4321:1234/1234",
			e:   "c1::4321:1234",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pf, err := port.ParsePF(u.exp)
			assert.Nil(t, err)
			assert.Equal(t, u.e, pf.String())
		})
	}
}
