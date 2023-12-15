// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRbacCmd(t *testing.T) {
	uu := map[string]struct {
		cmd  string
		ok   bool
		args []string
	}{
		"empty": {},
		"user": {
			cmd:  "can u:fernand",
			ok:   true,
			args: []string{"u", "fernand"},
		},
		"user_spacing": {
			cmd:  "can   u:  fernand  ",
			ok:   true,
			args: []string{"u", "fernand"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := newCmdParser(u.cmd)
			assert.Equal(t, u.ok, p.isRbacCmd())

			c, s, ok := p.parseRbac()
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.args[0], c)
				assert.Equal(t, u.args[1], s)
			}
		})
	}
}

func TestNsCmd(t *testing.T) {
	uu := map[string]struct {
		cmd string
		ok  bool
		ns  string
	}{
		"empty": {},
		"plain": {
			cmd: "pod fred",
			ok:  true,
			ns:  "fred",
		},
		"ns-arg": {
			cmd: "pod -n fred",
			ok:  true,
			ns:  "fred",
		},
		"ns-arg-spaced": {
			cmd: "pod   -n   fred   ",
			ok:  true,
			ns:  "fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := newCmdParser(u.cmd)
			ns, ok := p.nsCmd()
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.ns, ns)
			}
		})
	}
}

func TestFilterCmd(t *testing.T) {
	uu := map[string]struct {
		cmd    string
		ok     bool
		filter string
	}{
		"empty": {},
		"plain": {
			cmd:    "pod -ffred",
			ok:     true,
			filter: "fred",
		},
		"f-arg": {
			cmd:    "pod -f fred",
			ok:     true,
			filter: "fred",
		},
		"f-arg-spaced": {
			cmd:    "pod   -f   fred   ",
			ok:     true,
			filter: "fred",
		},
		"f-arg-caps": {
			cmd:    "POD   -F   FRED   ",
			ok:     true,
			filter: "fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := newCmdParser(u.cmd)
			f, ok := p.filterCmd()
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.filter, f)
			}
		})
	}
}

func TestLabelCmd(t *testing.T) {
	uu := map[string]struct {
		cmd    string
		ok     bool
		labels map[string]string
	}{
		"empty": {},
		"plain": {
			cmd:    "pod -lfred=blee",
			ok:     true,
			labels: map[string]string{"fred": "blee"},
		},
		"multi": {
			cmd:    "pod -l fred=blee,zorg=duh",
			ok:     true,
			labels: map[string]string{"fred": "blee", "zorg": "duh"},
		},
		"l-arg-spaced": {
			cmd:    "pod   -l   fred=blee   ",
			ok:     true,
			labels: map[string]string{"fred": "blee"},
		},
		"l-arg-caps": {
			cmd:    "POD   -L   FRED=BLEE   ",
			ok:     true,
			labels: map[string]string{"fred": "blee"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := newCmdParser(u.cmd)
			ll, ok := p.labelsCmd()
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.labels, ll)
			}
		})
	}
}
