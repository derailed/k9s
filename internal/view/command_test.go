package view

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// Test_viewMetaForAuthError ensures that when a command can't be resolved and
// the api server rejected us for auth reasons, we surface an authentication
// error instead of the misleading "command not found" (issue #3730).
func Test_viewMetaForAuthError(t *testing.T) {
	app := NewApp(mock.NewMockConfig(t))
	authErr := apierrors.NewUnauthorized("the server has asked for the client to provide credentials")
	app.Config.SetConnection(mock.NewMockConnectionWithError(authErr))

	c := &Command{
		app: app,
		alias: &dao.Alias{
			Aliases: config.NewAliases(),
		},
	}

	_, _, _, err := c.viewMetaFor(cmd.NewInterpreter("v1/pods"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
	assert.NotContains(t, err.Error(), "command not found")
	assert.ErrorIs(t, err, authErr)
}

// Test_viewMetaForNoAuthError ensures a genuinely unknown command still
// reports "command not found" when the connection is healthy.
func Test_viewMetaForNoAuthError(t *testing.T) {
	app := NewApp(mock.NewMockConfig(t))
	app.Config.SetConnection(mock.NewMockConnection())

	c := &Command{
		app: app,
		alias: &dao.Alias{
			Aliases: config.NewAliases(),
		},
	}

	_, _, _, err := c.viewMetaFor(cmd.NewInterpreter("v1/bogus"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command not found")
}

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
