package view

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
)

func Test_viewMetaFor(t *testing.T) {
	uu := map[string]struct {
		cmd string
		gvr *client.GVR
		p   *cmd.Interpreter
		err error
	}{
		"empty": {
			cmd: "",
			gvr: client.PodGVR,
			err: errors.New("`` command not found"),
		},

		"toast": {
			cmd: "v1/pd",
			gvr: client.PodGVR,
			err: errors.New("`v1/pd` command not found"),
		},

		"gvr": {
			cmd: "v1/pods",
			gvr: client.PodGVR,
			p:   cmd.NewInterpreter("v1/pods"),
			err: errors.New("blah"),
		},

		"short-name": {
			cmd: "po",
			gvr: client.PodGVR,
			p:   cmd.NewInterpreter("v1/pods", "po"),
			err: errors.New("blee"),
		},

		"custom-alias": {
			cmd: "pdl",
			gvr: client.PodGVR,
			p:   cmd.NewInterpreter("v1/pods @fred 'app=blee' default", "pdl"),
			err: errors.New("blee"),
		},

		"inception": {
			cmd: "pdal blee",
			gvr: client.PodGVR,
			p:   cmd.NewInterpreter("v1/pods @fred 'app=blee' blee", "pdal", "pod"),
			err: errors.New("blee"),
		},
	}

	c := &Command{
		alias: &dao.Alias{
			Aliases: config.NewAliases(),
		},
	}
	c.alias.Define(client.PodGVR, "po", "pod", "pods", client.PodGVR.String())
	c.alias.Define(client.NewGVR("pod default"), "pd")
	c.alias.Define(client.NewGVR("pod @fred 'app=blee' default"), "pdl")
	c.alias.Define(client.NewGVR("pdl"), "pdal")

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.cmd)
			gvr, _, acmd, err := c.viewMetaFor(p)
			if err != nil {
				assert.Equal(t, u.err.Error(), err.Error())
			} else {
				assert.Equal(t, u.gvr, gvr)
				assert.Equal(t, u.p, acmd)
			}
		})
	}
}

type mockClientWithInvalidate struct {
	client.Connection
	invalidated  bool
	onInvalidate func()
}

func (m *mockClientWithInvalidate) InvalidateCache() error {
	m.invalidated = true
	if m.onInvalidate != nil {
		m.onInvalidate()
	}
	return nil
}

func Test_viewMetaFor_MissingResourceRefreshes(t *testing.T) {
	oldMeta := dao.MetaAccess
	dao.MetaAccess = dao.NewMeta()
	defer func() {
		dao.MetaAccess = oldMeta
	}()

	clt := &mockClientWithInvalidate{}
	f := watch.NewFactory(clt)
	app := &App{factory: f}
	c := &Command{
		app: app,
		alias: &dao.Alias{
			Aliases: config.NewAliases(),
		},
	}

	p := cmd.NewInterpreter("hr")
	_, _, _, err := c.viewMetaFor(p)
	assert.Error(t, err)
	assert.Equal(t, "`hr` command not found", err.Error())
	assert.True(t, clt.invalidated)
}

func Test_viewMetaFor_MissingResourceDiscovered(t *testing.T) {
	oldMeta := dao.MetaAccess
	dao.MetaAccess = dao.NewMeta()
	defer func() {
		dao.MetaAccess = oldMeta
	}()

	hrGVR := client.NewGVR("helm.toolkit.fluxcd.io/v2beta1/helmreleases")
	clt := &mockClientWithInvalidate{}
	f := watch.NewFactory(clt)
	app := &App{factory: f}
	c := &Command{
		app: app,
		alias: &dao.Alias{
			Aliases: config.NewAliases(),
		},
	}
	clt.onInvalidate = func() {
		c.alias.Define(hrGVR, "hr", "helmrelease")
	}

	p := cmd.NewInterpreter("hr")
	gvr, _, _, err := c.viewMetaFor(p)
	assert.NoError(t, err)
	assert.Equal(t, hrGVR, gvr)
	assert.True(t, clt.invalidated)
}
