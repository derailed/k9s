// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package watch_test

import (
	"testing"
	"time"

	"github.com/derailed/k9s/internal/port"
	"github.com/derailed/k9s/internal/watch"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/portforward"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestIsPodForwarded(t *testing.T) {
	uu := map[string]struct {
		ff  watch.Forwarders
		fqn string
		e   bool
	}{
		"happy": {
			ff: watch.Forwarders{
				"ns1/p1||8080:8080": newNoOpForwarder(),
			},
			fqn: "ns1/p1",
			e:   true,
		},
		"dud": {
			ff: watch.Forwarders{
				"ns1/p1||8080:8080": newNoOpForwarder(),
			},
			fqn: "ns1/p2",
		},
		"sub": {
			ff: watch.Forwarders{
				"ns1/freddy||8080:8080": newNoOpForwarder(),
			},
			fqn: "ns1/fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.ff.IsPodForwarded(u.fqn))
		})
	}
}

func TestIsContainerForwarded(t *testing.T) {
	uu := map[string]struct {
		ff      watch.Forwarders
		fqn, co string
		e       bool
	}{
		"happy": {
			ff: watch.Forwarders{
				"ns1/p1|c1|8080:8080": newNoOpForwarder(),
			},
			fqn: "ns1/p1",
			co:  "c1",
			e:   true,
		},
		"dud": {
			ff: watch.Forwarders{
				"ns1/p1|c1|8080:8080": newNoOpForwarder(),
			},
			fqn: "ns1/p1",
			co:  "c2",
		},
		"sub": {
			ff: watch.Forwarders{
				"ns1/freddy|c1|8080:8080": newNoOpForwarder(),
			},
			fqn: "ns1/fred",
			co:  "c1",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.ff.IsContainerForwarded(u.fqn, u.co))
		})
	}
}

func TestKill(t *testing.T) {
	uu := map[string]struct {
		ff    watch.Forwarders
		path  string
		kills int
	}{
		"partial_match": {
			ff: watch.Forwarders{
				"ns1/p1|c1|8080:8080":   newNoOpForwarder(),
				"ns1/p1_1|c1|8080:8080": newNoOpForwarder(),
				"ns1/p2|c1|8080:8080":   newNoOpForwarder(),
			},
			path:  "ns1/p1",
			kills: 1,
		},
		"partial_no_match": {
			ff: watch.Forwarders{
				"ns1/p1|c1|8080:8080":   newNoOpForwarder(),
				"ns1/p1_1|c1|8080:8080": newNoOpForwarder(),
				"ns1/p2|c1|8080:8080":   newNoOpForwarder(),
			},
			path: "ns1/p",
		},
		"path_sub": {
			ff: watch.Forwarders{
				"ns1/p1|c1|8080:8080":   newNoOpForwarder(),
				"ns1/p1_1|c1|8080:8080": newNoOpForwarder(),
				"ns1/p2|c1|8080:8080":   newNoOpForwarder(),
			},
			path:  "ns1/p1",
			kills: 1,
		},
		"partial_multi": {
			ff: watch.Forwarders{
				"ns1/p1|c1|8080:8080": newNoOpForwarder(),
				"ns1/p1|c2|8081:8081": newNoOpForwarder(),
				"ns1/p2|c1|8080:8080": newNoOpForwarder(),
			},
			path:  "ns1/p1",
			kills: 2,
		},
		"full_match": {
			ff: watch.Forwarders{
				"ns1/p1|c1|8080:8080":   newNoOpForwarder(),
				"ns1/p1_1|c1|8080:8080": newNoOpForwarder(),
				"ns1/p2|c1|8080:8080":   newNoOpForwarder(),
			},
			path:  "ns1/p1|c1|8080:8080",
			kills: 1,
		},
		"full_no_match_co": {
			ff: watch.Forwarders{
				"ns1/p1|c1|8080:8080":   newNoOpForwarder(),
				"ns1/p1_1|c1|8080:8080": newNoOpForwarder(),
				"ns1/p2|c1|8080:8080":   newNoOpForwarder(),
			},
			path: "ns1/p1|c2|8080:8080",
		},
		"full_no_match_ports": {
			ff: watch.Forwarders{
				"ns1/p1|c1|8080:8080":   newNoOpForwarder(),
				"ns1/p1_1|c1|8080:8080": newNoOpForwarder(),
				"ns1/p2|c1|8080:8080":   newNoOpForwarder(),
			},
			path: "ns1/p1|c1|8081:8080",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.kills, u.ff.Kill(u.path))
		})
	}
}

type noOpForwarder struct{}

func newNoOpForwarder() noOpForwarder {
	return noOpForwarder{}
}

func (noOpForwarder) Start(path string, tunnel port.PortTunnel) (*portforward.PortForwarder, error) {
	return nil, nil
}
func (noOpForwarder) Stop()                      {}
func (noOpForwarder) ID() string                 { return "" }
func (noOpForwarder) Container() string          { return "" }
func (noOpForwarder) Port() string               { return "" }
func (noOpForwarder) FQN() string                { return "" }
func (noOpForwarder) Active() bool               { return false }
func (noOpForwarder) SetActive(bool)             {}
func (noOpForwarder) Age() time.Time             { return time.Now() }
func (noOpForwarder) HasPortMapping(string) bool { return false }
func (noOpForwarder) Address() string            { return "" }
