// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port_test

import (
	"testing"

	"github.com/derailed/k9s/internal/port"
	"github.com/stretchr/testify/assert"
)

func TestPortTunnelMap(t *testing.T) {
	uu := map[string]struct {
		pt              port.PortTunnel
		coPort, locPort string
		e               string
	}{
		"plain": {
			pt: port.PortTunnel{
				Address:       "localhost",
				LocalPort:     "1234",
				ContainerPort: "4321",
			},
			e: "1234:4321",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.pt.PortMap())
		})
	}
}
