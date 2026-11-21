// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package port_test

import (
	"context"
	"net"
	"testing"

	"github.com/derailed/k9s/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestIsPortFree(t *testing.T) {
	uu := map[string]string{
		"ipv4":     "127.0.0.1",
		"ipv6":     "::1",
		"hostname": "localhost",
	}

	for k, addr := range uu {
		t.Run(k, func(t *testing.T) {
			var lc net.ListenConfig
			l, err := lc.Listen(context.Background(), "tcp", net.JoinHostPort(addr, "0"))
			if err != nil {
				t.Skipf("no loopback listener on %s: %v", addr, err)
			}
			_, p, err := net.SplitHostPort(l.Addr().String())
			require.NoError(t, err)
			pt := port.NewPortTunnel(addr, "c1", p, "80")

			assert.False(t, port.IsPortFree(context.Background(), pt))
			require.NoError(t, l.Close())
			assert.True(t, port.IsPortFree(context.Background(), pt))
		})
	}
}
