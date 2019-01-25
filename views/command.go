package views

import (
	"fmt"

	"github.com/k8sland/k9s/resource"
)

type command struct {
	app *appView
}

func newCommand(app *appView) *command {
	return &command{app: app}
}

// DefaultCmd reset default command ie show pods.
func (c *command) defaultCmd() {
	c.run("po")
}

// Helpers...

// Exec the command by showing associated display.
func (c *command) run(cmd string) {
	var v igniter
	switch cmd {
	case "q":
		c.app.quit(nil)
	case "h", "help", "?":
		v = newHelpView(c.app)
		c.app.flash(flashInfo, "Viewing Help...")
	default:
		if res, ok := cmdMap[cmd]; ok {
			v = res.viewFn(res.title, c.app, res.listFn(defaultNS), res.colorerFn)
			c.app.flash(flashInfo, "Viewing all "+res.title+"...")
		} else {
			if res, ok := getCRDS()[cmd]; ok {
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
			} else {
				c.app.flash(flashWarn, fmt.Sprintf("Huh? `%s` command not found", cmd))
			}
		}
	}

	if v != nil {
		c.app.inject(v)
	}
	return
}
