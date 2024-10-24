// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Node represents a node view.
type Node struct {
	ResourceViewer
}

// NewNode returns a new node view.
func NewNode(gvr client.GVR) ResourceViewer {
	n := Node{
		ResourceViewer: NewBrowser(gvr),
	}
	n.AddBindKeysFn(n.bindKeys)
	n.GetTable().SetEnterFn(n.showPods)
	n.SetContextFn(n.nodeContext)

	return &n
}

func (n *Node) nodeContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyPodCounting, !n.App().Config.K9s.DisablePodCounting)
}

func (n *Node) bindDangerousKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyC: ui.NewKeyActionWithOpts(
			"Cordon",
			n.toggleCordonCmd(true),
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			},
		),
		ui.KeyU: ui.NewKeyActionWithOpts(
			"Uncordon",
			n.toggleCordonCmd(false),
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			},
		),
		ui.KeyR: ui.NewKeyActionWithOpts(
			"Drain",
			n.drainCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			},
		),
	})
	ct, err := n.App().Config.K9s.ActiveContext()
	if err != nil {
		log.Error().Err(err).Msgf("No active context located")
		return
	}
	if ct.FeatureGates.NodeShell {
		aa.Add(ui.KeyS, ui.NewKeyAction("Shell", n.sshCmd, true))
	}
}

func (n *Node) bindKeys(aa *ui.KeyActions) {
	if !n.App().Config.K9s.IsReadOnly() {
		n.bindDangerousKeys(aa)
	}

	aa.Bulk(ui.KeyMap{
		ui.KeyY:      ui.NewKeyAction(yamlAction, n.yamlCmd, true),
		ui.KeyShiftR: ui.NewKeyAction("Sort ROLE", n.GetTable().SortColCmd("ROLE", true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort CPU", n.GetTable().SortColCmd(cpuCol, false), false),
		ui.KeyShiftM: ui.NewKeyAction("Sort MEM", n.GetTable().SortColCmd(memCol, false), false),
		ui.KeyShiftO: ui.NewKeyAction("Sort Pods", n.GetTable().SortColCmd("PODS", false), false),
	})
}

func (n *Node) showPods(a *App, _ ui.Tabular, _ client.GVR, path string) {
	showPods(a, n.GetTable().GetSelectedItem(), client.BlankNamespace, "spec.nodeName="+path)
}

func (n *Node) drainCmd(evt *tcell.EventKey) *tcell.EventKey {
	sels := n.GetTable().GetSelectedItems()
	if len(sels) == 0 {
		return evt
	}

	opts := dao.DrainOptions{
		GracePeriodSeconds: -1,
		Timeout:            5 * time.Second,
	}
	ShowDrain(n, sels, opts, drainNode)

	return nil
}

func drainNode(v ResourceViewer, sels []string, opts dao.DrainOptions) {
	res, err := dao.AccessorFor(v.App().factory, v.GVR())
	if err != nil {
		v.App().Flash().Err(err)
		return
	}
	m, ok := res.(dao.NodeMaintainer)
	if !ok {
		v.App().Flash().Err(fmt.Errorf("expecting a maintainer for %q", v.GVR()))
		return
	}

	v.Stop()
	defer v.Start()
	{
		d := NewDetails(v.App(), "Drain Progress", "nodes", contentYAML, true)
		if err := v.App().inject(d, false); err != nil {
			v.App().Flash().Err(err)
		}
		for _, sel := range sels {
			if err := m.Drain(sel, opts, d.GetWriter()); err != nil {
				v.App().Flash().Err(err)
			}
		}
		v.Refresh()
	}
}

func (n *Node) toggleCordonCmd(cordon bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		sels := n.GetTable().GetSelectedItems()
		if len(sels) == 0 {
			return evt
		}

		title, msg := "Confirm ", ""
		if cordon {
			title, msg = title+"Cordon", "Cordon "
		} else {
			title, msg = title+"Uncordon", "Uncordon "
		}
		if len(sels) == 1 {
			msg += sels[0] + "?"
		} else {
			msg += fmt.Sprintf("(%d) marked %s?", len(sels), n.GVR().R())
		}
		dialog.ShowConfirm(n.App().Styles.Dialog(), n.App().Content.Pages, title, msg, func() {
			res, err := dao.AccessorFor(n.App().factory, n.GVR())
			if err != nil {
				n.App().Flash().Err(err)
				return
			}
			m, ok := res.(dao.NodeMaintainer)
			if !ok {
				n.App().Flash().Err(fmt.Errorf("expecting a maintainer for %q", n.GVR()))
				return
			}
			for _, s := range sels {
				if err := m.ToggleCordon(s, cordon); err != nil {
					n.App().Flash().Err(err)
				}
			}
			n.Refresh()
		}, func() {})

		return nil
	}
}

func (n *Node) sshCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := n.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	n.Stop()
	defer n.Start()
	_, node := client.Namespaced(path)
	launchNodeShell(n, n.App(), node)

	return nil
}

func (n *Node) yamlCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := n.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	n.Stop()
	defer n.Start()
	ctx, cancel := context.WithTimeout(context.Background(), n.App().Conn().Config().CallTimeout())
	defer cancel()

	sel := n.GetTable().GetSelectedItem()
	gvr := n.GVR().GVR()
	dial, err := n.App().factory.Client().DynDial()
	if err != nil {
		n.App().Flash().Err(err)
		return nil
	}
	o, err := dial.Resource(gvr).Get(ctx, sel, metav1.GetOptions{})
	if err != nil {
		n.App().Flash().Errf("Unable to get resource %q -- %s", n.GVR(), err)
		return nil
	}

	raw, err := dao.ToYAML(o, false)
	if err != nil {
		n.App().Flash().Errf("Unable to marshal resource %s", err)
		return nil
	}

	details := NewDetails(n.App(), yamlAction, sel, contentYAML, true).Update(raw)
	if err := n.App().inject(details, false); err != nil {
		n.App().Flash().Err(err)
	}

	return nil
}
