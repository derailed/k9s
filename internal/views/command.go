package views

import (
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/rs/zerolog/log"
)

type command struct {
	app     *appView
	history *cmdStack
}

func newCommand(app *appView) *command {
	return &command{app: app, history: newCmdStack()}
}

func (c *command) pushCmd(cmd string) {
	c.history.push(cmd)
	c.app.crumbsView.update(c.history.stack)
}

func (c *command) previousCmd() (string, bool) {
	c.history.pop()
	c.app.crumbsView.update(c.history.stack)
	return c.history.top()
}

// DefaultCmd reset default command ie show pods.
func (c *command) defaultCmd() {
	c.pushCmd(c.app.config.ActiveView())
	c.run(c.app.config.ActiveView())
}

// Helpers...

// Exec the command by showing associated display.
func (c *command) run(cmd string) bool {
	defer func() {
		if err := recover(); err != nil {
			log.Debug().Msgf("Command failed %v", err)
		}
	}()

	var v resourceViewer
	switch cmd {
	case "q", "quit":
		c.app.Stop()
		return true
	case "?", "help", "alias":
		c.app.inject(newAliasView(c.app))
		return true
	default:
		if res, ok := resourceViews()[cmd]; ok {
			var r resource.List
			if res.listMxFn != nil {
				r = res.listMxFn(c.app.conn(),
					k8s.NewMetricsServer(c.app.conn()),
					resource.DefaultNamespace,
				)
			} else {
				r = res.listFn(c.app.conn(), resource.DefaultNamespace)
			}
			v = res.viewFn(res.title, c.app, r, res.colorerFn)
			if res.enterFn != nil {
				v.setEnterFn(res.enterFn)
			}
			const fmat = "Viewing %s in namespace %s..."
			c.app.flash(flashInfo, fmt.Sprintf(fmat, res.title, c.app.config.ActiveNamespace()))
			log.Debug().Msgf("Running command %s", cmd)
			c.exec(cmd, v)
			return true
		}
	}

	res, ok := allCRDs(c.app.conn())[cmd]
	if !ok {
		c.app.flash(flashWarn, fmt.Sprintf("Huh? `%s` command not found", cmd))
		return false
	}

	n := res.Plural
	if len(n) == 0 {
		n = res.Singular
	}
	v = newResourceView(
		res.Kind,
		c.app,
		resource.NewCustomList(c.app.conn(), "", res.Group, res.Version, n),
		defaultColorer,
	)
	c.exec(cmd, v)
	return true
}

func (c *command) exec(cmd string, v igniter) {
	if v != nil {
		c.app.config.SetActiveView(cmd)
		c.app.config.Save()
		c.app.inject(v)
	}
}
