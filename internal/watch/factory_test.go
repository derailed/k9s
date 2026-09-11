// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package watch_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spyConn records the namespaces CanI is asked about and answers from an allow-list.
type spyConn struct {
	client.Connection
	seen  []string
	allow map[string]bool
}

func (c *spyConn) CanI(ns string, _ *client.GVR, _ string, _ []string) (bool, error) {
	c.seen = append(c.seen, ns)

	return c.allow[ns], nil
}

func TestFactoryCanForResourceMultiNS(t *testing.T) {
	// Deny every individual namespace but "allow" the joined string. The fix must
	// decompose "ns1,ns2" and check each namespace on its own, so it never queries
	// the joined value and access ends up denied for all -> error.
	conn := &spyConn{
		Connection: mock.NewMockConnection(),
		allow:      map[string]bool{"ns1,ns2": true},
	}
	f := watch.NewFactory(conn)

	_, err := f.CanForResource("ns1,ns2", client.PodGVR, client.ListAccess)
	require.Error(t, err)
	assert.Equal(t, []string{"ns1", "ns2"}, conn.seen)
}
