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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestBrowser() ResourceViewer {
	b := NewBrowser(client.PodGVR).(*Browser)
	b.meta = &metav1.APIResource{Kind: "Pod"}

	return NewScaleExtender(NewOwnerExtender(b))
}

func TestFilterCmd(t *testing.T) {
	uu := map[string]struct {
		cmd    string
		filter string
	}{
		"no-filter": {
			cmd: "filter",
		},
		"with-arg": {
			cmd:    "filter !Completed",
			filter: "!completed",
		},
		"with-slash-arg": {
			cmd:    "filter /fred",
			filter: "fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			a := NewApp(mock.NewMockConfig(t))
			comp := newTestBrowser()
			a.Content.Push(comp)

			c := &Command{app: a}
			p := cmd.NewInterpreter(u.cmd)
			c.filterCmd(p)

			buff := comp.(TableViewer).GetTable().CmdBuff()
			if u.filter != "" {
				assert.Equal(t, u.filter, buff.GetText())
				assert.False(t, buff.IsActive())
			} else {
				assert.True(t, buff.IsActive())
			}
		})
	}
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
