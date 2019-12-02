package view

import (
	"regexp"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	favNSIndicator     = "+"
	defaultNSIndicator = "(*)"
)

var nsCleanser = regexp.MustCompile(`(\w+)[+|(*)|(𝜟)]*`)

// Namespace represents a namespace viewer.
type Namespace struct {
	ResourceViewer
}

// NewNamespace returns a new viewer
func NewNamespace(gvr dao.GVR) ResourceViewer {
	n := Namespace{
		ResourceViewer: NewGeneric(gvr),
	}
	n.GetTable().SetDecorateFn(n.decorate)
	n.GetTable().SetColorerFn(render.Namespace{}.ColorerFunc())
	n.GetTable().SetEnterFn(n.switchNs)
	n.GetTable().SetSelectedFn(n.cleanser)
	n.BindKeys()

	return &n
}

func (n *Namespace) BindKeys() {
	n.Actions().Add(ui.KeyActions{
		ui.KeyU: ui.NewKeyAction("Use", n.useNsCmd, true),
	})
}

func (n *Namespace) switchNs(app *App, _, res, sel string) {
	n.useNamespace(sel)
	app.gotoResource("po")
}

func (n *Namespace) useNsCmd(evt *tcell.EventKey) *tcell.EventKey {
	ns := n.GetTable().GetSelectedItem()
	if ns == "" {
		return evt
	}
	n.useNamespace(ns)

	return nil
}

func (n *Namespace) useNamespace(ns string) {
	if err := n.App().Config.SetActiveNamespace(ns); err != nil {
		n.App().Flash().Err(err)
	} else {
		n.App().Flash().Infof("Namespace %s is now active!", ns)
	}
	if err := n.App().Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config file save failed!")
	}
	n.App().switchNS(ns)
}

func (*Namespace) cleanser(s string) string {
	log.Debug().Msgf("NS CLEANZ %q", s)
	return nsCleanser.ReplaceAllString(s, `$1`)
}

func (n *Namespace) decorate(data render.TableData) render.TableData {
	if n.App().Conn() == nil {
		return render.TableData{}
	}

	log.Debug().Msgf("CLONING %q", data.Namespace)
	// don't want to change the cache here thus need to clone!!
	res := data.Clone()
	// checks if all ns is in the list if not add it.
	if _, ok := data.RowEvents.FindIndex(render.NamespaceAll); !ok {
		res.RowEvents = append(render.RowEvents{
			render.RowEvent{
				Kind: render.EventUnchanged,
				Row: render.Row{
					ID:     render.AllNamespaces,
					Fields: render.Fields{render.NamespaceAll, "Active", "0"},
				},
			},
		},
			res.RowEvents...)
	}

	for _, re := range res.RowEvents {
		if config.InList(n.App().Config.FavNamespaces(), re.Row.ID) {
			re.Row.Fields[0] += favNSIndicator
			re.Kind = render.EventUnchanged
		}
		if n.App().Config.ActiveNamespace() == re.Row.ID {
			re.Row.Fields[0] += defaultNSIndicator
			re.Kind = render.EventUnchanged
		}
	}

	return res
}
