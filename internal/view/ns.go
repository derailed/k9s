package view

import (
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	favNSIndicator     = "+"
	defaultNSIndicator = "(*)"
)

// Namespace represents a namespace viewer.
type Namespace struct {
	ResourceViewer
}

// NewNamespace returns a new viewer
func NewNamespace(gvr client.GVR) ResourceViewer {
	n := Namespace{
		ResourceViewer: NewBrowser(gvr),
	}
	n.GetTable().SetDecorateFn(n.decorate)
	n.GetTable().SetColorerFn(render.Namespace{}.ColorerFunc())
	n.GetTable().SetEnterFn(n.switchNs)
	n.SetBindKeysFn(n.bindKeys)

	return &n
}

func (n *Namespace) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyU: ui.NewKeyAction("Use", n.useNsCmd, true),
	})
}

func (n *Namespace) switchNs(app *App, model ui.Tabular, gvr, path string) {
	n.useNamespace(path)
	if err := app.gotoResource("pods", "", true); err != nil {
		app.Flash().Err(err)
	}
}

func (n *Namespace) useNsCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := n.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}
	n.useNamespace(path)

	return nil
}

func (n *Namespace) useNamespace(fqn string) {
	_, ns := client.Namespaced(fqn)
	log.Debug().Msgf("SWITCHING NS %q", ns)
	n.App().switchNS(ns)
	if err := n.App().Config.SetActiveNamespace(ns); err != nil {
		n.App().Flash().Err(err)
	} else {
		n.App().Flash().Infof("Namespace %s is now active!", ns)
	}
	if err := n.App().Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config file save failed!")
	}
}

func (n *Namespace) decorate(data render.TableData) render.TableData {
	if n.App().Conn() == nil || len(data.RowEvents) == 0 {
		return data
	}

	// checks if all ns is in the list if not add it.
	if _, ok := data.RowEvents.FindIndex(client.NamespaceAll); !ok {
		log.Debug().Msg("YO!!")
		data.RowEvents = append(data.RowEvents,
			render.RowEvent{
				Kind: render.EventUnchanged,
				Row: render.Row{
					ID:     client.NamespaceAll,
					Fields: render.Fields{client.NamespaceAll, "Active", "", "", time.Now().String()},
				},
			},
		)
	}

	for _, re := range data.RowEvents {
		if config.InList(n.App().Config.FavNamespaces(), re.Row.ID) {
			re.Row.Fields[0] += favNSIndicator
			re.Kind = render.EventUnchanged
		}
		if n.App().Config.ActiveNamespace() == re.Row.ID {
			re.Row.Fields[0] += defaultNSIndicator
			re.Kind = render.EventUnchanged
		}
	}

	return data
}
