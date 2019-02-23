package views

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
)

type command struct {
	app *appView
}

func newCommand(app *appView) *command {
	return &command{app: app}
}

// DefaultCmd reset default command ie show pods.
func (c *command) defaultCmd() {
	c.run(config.Root.ActiveView())
}

// Helpers...

// Exec the command by showing associated display.
func (c *command) run(cmd string) {
	var v igniter
	switch cmd {
	case "q", "quit":
		c.app.Stop()
		return
	case "?", "help", "alias":
		c.app.inject(newAliasView(c.app))
		return
	default:
		if res, ok := cmdMap[cmd]; ok {
			v = res.viewFn(res.title, c.app, res.listFn(resource.DefaultNamespace), res.colorerFn)
			c.app.flash(flashInfo, "Viewing all "+res.title+"...")
			c.exec(cmd, v)
			return
		}
	}

	res, ok := getCRDS()[cmd]
	if !ok {
		c.app.flash(flashWarn, fmt.Sprintf("Huh? `%s` command not found", cmd))
		return
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
}

func (c *command) exec(cmd string, v igniter) {
	if v != nil {
		config.Root.SetActiveView(cmd)
		config.Root.Save()
		c.app.inject(v)
	}
}
