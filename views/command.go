package views

import (
	"fmt"

	"github.com/derailed/k9s/resource"
)

type command struct {
	app *appView
}

func newCommand(app *appView) *command {
	return &command{app: app}
}

// DefaultCmd reset default command ie show pods.
func (c *command) defaultCmd() {
	c.run(k9sCfg.K9s.View.Active)
}

// Helpers...

// Exec the command by showing associated display.
func (c *command) run(cmd string) {
	var v igniter
	if res, ok := cmdMap[cmd]; ok {
		v = res.viewFn(res.title, c.app, res.listFn(defaultNS), res.colorerFn)
		c.app.flash(flashInfo, "Viewing all "+res.title+"...")
		c.exec(cmd, v)
		return
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
		k9sCfg.K9s.View.Active = cmd
		k9sCfg.validateAndSave()
		c.app.inject(v)
	}
}
