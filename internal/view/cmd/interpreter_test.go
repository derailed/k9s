// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd_test

import (
	"testing"

	"github.com/derailed/k9s/internal/view/cmd"
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
			p := cmd.NewInterpreter(u.cmd)
			assert.Equal(t, u.ok, p.IsRBACCmd())

			c, s, ok := p.RBACArgs()
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
		"happy": {
			cmd: "pod fred",
			ok:  true,
			ns:  "fred",
		},
		"ns-arg-spaced": {
			cmd: "pod      fred   ",
			ok:  true,
			ns:  "fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			ns, ok := p.NSArg()
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
		"normal": {
			cmd:    "pod /fred",
			ok:     true,
			filter: "fred",
		},
		"caps": {
			cmd:    "POD /FRED",
			ok:     true,
			filter: "fred",
		},
		"filter+ns": {
			cmd:    "pod /fred ns1",
			ok:     true,
			filter: "fred",
		},
		"ns+filter": {
			cmd:    "pod ns1 /fred",
			ok:     true,
			filter: "fred",
		},
		"ns+filter+labels": {
			cmd:    "pod ns1 /fred app=blee,fred=zorg",
			ok:     true,
			filter: "fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			f, ok := p.FilterArg()
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
			cmd:    "pod fred=blee",
			ok:     true,
			labels: map[string]string{"fred": "blee"},
		},
		"multi": {
			cmd:    "pod fred=blee,zorg=duh",
			ok:     true,
			labels: map[string]string{"fred": "blee", "zorg": "duh"},
		},
		"multi-ns": {
			cmd:    "pod fred=blee,zorg=duh ns1",
			ok:     true,
			labels: map[string]string{"fred": "blee", "zorg": "duh"},
		},
		"l-arg-spaced": {
			cmd:    "pod   fred=blee   ",
			ok:     true,
			labels: map[string]string{"fred": "blee"},
		},
		"l-arg-caps": {
			cmd:    "POD  FRED=BLEE   ",
			ok:     true,
			labels: map[string]string{"fred": "blee"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			ll, ok := p.LabelsArg()
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.labels, ll)
			}
		})
	}
}

func TestXRayCmd(t *testing.T) {
	uu := map[string]struct {
		cmd     string
		ok      bool
		res, ns string
	}{
		"empty": {},

		"happy": {
			cmd: "xray po",
			ok:  true,
			res: "po",
		},

		"happy+ns": {
			cmd: "xray po ns1",
			ok:  true,
			res: "po",
			ns:  "ns1",
		},

		"toast": {
			cmd: "xrayzor po",
		},

		"toast-1": {
			cmd: "xray",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			res, ns, ok := p.XrayArgs()
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.res, res)
				assert.Equal(t, u.ns, ns)
			}
		})
	}
}

func TestDirCmd(t *testing.T) {
	uu := map[string]struct {
		cmd string
		ok  bool
		dir string
	}{
		"empty": {},

		"happy": {
			cmd: "dir dir1",
			ok:  true,
			dir: "dir1",
		},

		"extra-ns": {
			cmd: "dir dir1 ns1",
			ok:  true,
			dir: "dir1",
		},

		"toast": {
			cmd: "dirdel dir1",
		},

		"toast-nodir": {
			cmd: "dir",
		},
		"caps": {
			cmd: "dir DirName",
			ok:  true,
			dir: "DirName",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			dir, ok := p.DirArg()
			assert.Equal(t, u.ok, ok)
			assert.Equal(t, u.dir, dir)
		})
	}
}

func TestRBACCmd(t *testing.T) {
	uu := map[string]struct {
		cmd      string
		ok       bool
		cat, sub string
	}{
		"empty": {},
		"toast": {
			cmd: "canopy u:bozo",
		},
		"toast-1": {
			cmd: "can u:",
		},
		"toast-2": {
			cmd: "can bozo",
		},
		"user": {
			cmd: "can u:bozo",
			ok:  true,
			cat: "u",
			sub: "bozo",
		},
		"group": {
			cmd: "can g:bozo",
			ok:  true,
			cat: "g",
			sub: "bozo",
		},
		"sa": {
			cmd: "can s:bozo",
			ok:  true,
			cat: "s",
			sub: "bozo",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			cat, sub, ok := p.RBACArgs()
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.cat, cat)
				assert.Equal(t, u.sub, sub)
			}
		})
	}
}

func TestContextCmd(t *testing.T) {
	uu := map[string]struct {
		cmd string
		ok  bool
		ctx string
	}{
		"empty": {},
		"happy-full": {
			cmd: "context ctx1",
			ok:  true,
			ctx: "ctx1",
		},
		"happy-alias": {
			cmd: "ctx ctx1",
			ok:  true,
			ctx: "ctx1",
		},
		"toast": {
			cmd: "ctxto ctx1",
		},
		"caps": {
			cmd: "ctx Dev",
			ok:  true,
			ctx: "Dev",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			assert.Equal(t, u.ok, p.IsContextCmd())
			if u.ok {
				ctx, ok := p.ContextArg()
				assert.Equal(t, u.ok, ok)
				assert.Equal(t, u.ctx, ctx)
			}
		})
	}
}

func TestHelpCmd(t *testing.T) {
	uu := map[string]struct {
		cmd string
		ok  bool
	}{
		"empty": {},
		"plain": {
			cmd: "help",
			ok:  true,
		},
		"toast": {
			cmd: "helpme",
		},
		"toast1": {
			cmd: "hozer",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			assert.Equal(t, u.ok, p.IsHelpCmd())
		})
	}
}

func TestBailCmd(t *testing.T) {
	uu := map[string]struct {
		cmd string
		ok  bool
	}{
		"empty": {},
		"plain": {
			cmd: "quit",
			ok:  true,
		},
		"q": {
			cmd: "q",
			ok:  true,
		},
		"q!": {
			cmd: "q!",
			ok:  true,
		},
		"toast": {
			cmd: "zorg",
		},
		"toast1": {
			cmd: "quitter",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			assert.Equal(t, u.ok, p.IsBailCmd())
		})
	}
}

func TestAliasCmd(t *testing.T) {
	uu := map[string]struct {
		cmd string
		ok  bool
	}{
		"empty": {},
		"plain": {
			cmd: "alias",
			ok:  true,
		},
		"a": {
			cmd: "a",
			ok:  true,
		},
		"toast": {
			cmd: "abba",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			assert.Equal(t, u.ok, p.IsAliasCmd())
		})
	}
}

func TestCowCmd(t *testing.T) {
	uu := map[string]struct {
		cmd string
		ok  bool
	}{
		"empty": {},
		"plain": {
			cmd: "cow",
			ok:  true,
		},
		"msg": {
			cmd: "cow bumblebeetuna",
			ok:  true,
		},
		"toast": {
			cmd: "cowdy",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			assert.Equal(t, u.ok, p.IsCowCmd())
		})
	}
}
