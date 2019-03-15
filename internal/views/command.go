package views

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
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
	c.pushCmd(config.Root.ActiveView())
	c.run(config.Root.ActiveView())
}

// Helpers...

// Exec the command by showing associated display.
func (c *command) run(cmd string) bool {
	var v igniter
	switch cmd {
	case "q", "quit":
		c.app.Stop()
		return true
	case "?", "help", "alias":
		c.app.inject(newAliasView(c.app))
		return true
	default:
		if res, ok := resourceViews()[cmd]; ok {
			v = res.viewFn(res.title, c.app, res.listFn(resource.DefaultNamespace), res.colorerFn)
			c.app.flash(flashInfo, fmt.Sprintf("Viewing %s in namespace %s...", res.title, config.Root.ActiveNamespace()))
			c.exec(cmd, v)
			return true
		}
	}

	res, ok := allCRDs()[cmd]
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
		resource.NewCustomList("", res.Group, res.Version, n),
		defaultColorer,
	)
	c.exec(cmd, v)
	return true
}

func (c *command) exec(cmd string, v igniter) {
	if v != nil {
		config.Root.SetActiveView(cmd)
		config.Root.Save()
		c.app.inject(v)
	}
}
