package views

import (
	"fmt"
	"regexp"

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

var fuMatcher = regexp.MustCompile(`\Afu\s([u|g|s]):([\w-:]+)\b`)

// Exec the command by showing associated display.
func (c *command) run(cmd string) bool {
	var v resourceViewer
	switch {
	case cmd == "q", cmd == "quit":
		c.app.Stop()
		return true
	case cmd == "?", cmd == "help":
		c.app.inject(newHelpView(c.app))
		return true
	case cmd == "alias":
		c.app.inject(newAliasView(c.app))
		return true
	case fuMatcher.MatchString(cmd):
		tokens := fuMatcher.FindAllStringSubmatch(cmd, -1)
		if len(tokens) == 1 && len(tokens[0]) == 3 {
			c.app.inject(newFuView(c.app, tokens[0][1], tokens[0][2]))
			return true
		}
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
			v = res.viewFn(res.title, c.app, r)
			if res.colorerFn != nil {
				v.setColorerFn(res.colorerFn)
			}
			if res.enterFn != nil {
				v.setEnterFn(res.enterFn)
			}
			if res.decorateFn != nil {
				v.setDecorateFn(res.decorateFn)
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

	name := res.Plural
	if len(name) == 0 {
		name = res.Singular
	}
	v = newResourceView(
		res.Kind,
		c.app,
		resource.NewCustomList(c.app.conn(), "", res.Group, res.Version, name),
	)
	v.setColorerFn(defaultColorer)
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
