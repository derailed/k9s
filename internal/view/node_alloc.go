// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeAlloc represents a node allocation view that always shows allocated resources.
type NodeAlloc struct {
	ResourceViewer
}

// NewNodeAlloc returns a new node allocation view.
// This view always enables pod counting to show allocated CPU and memory.
func NewNodeAlloc(gvr *client.GVR) ResourceViewer {
	n := NodeAlloc{
		ResourceViewer: NewBrowser(gvr),
	}
	n.AddBindKeysFn(n.bindKeys)
	n.GetTable().SetEnterFn(n.showPods)
	n.SetContextFn(n.nodeAllocContext)

	return &n
}

func (n *NodeAlloc) nodeAllocContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyPodCounting, true)
}

func (n *NodeAlloc) bindDangerousKeys(aa *ui.KeyActions) {
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
		slog.Error("No active context located", slogs.Error, err)
		return
	}
	if ct.FeatureGates.NodeShell && n.App().Config.K9s.ShellPod != nil {
		aa.Add(ui.KeyS, ui.NewKeyAction("Shell", n.sshCmd, true))
	}
}

func (n *NodeAlloc) bindKeys(aa *ui.KeyActions) {
	if !n.App().Config.IsReadOnly() {
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

func (n *NodeAlloc) showPods(a *App, _ ui.Tabular, _ *client.GVR, path string) {
	showPods(a, n.GetTable().GetSelectedItem(), nil, "spec.nodeName="+path)
}

func (n *NodeAlloc) drainCmd(evt *tcell.EventKey) *tcell.EventKey {
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

func (n *NodeAlloc) toggleCordonCmd(cordon bool) func(evt *tcell.EventKey) *tcell.EventKey {
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
		d := n.App().Styles.Dialog()
		dialog.ShowConfirm(&d, n.App().Content.Pages, title, msg, func() {
			res, err := dao.AccessorFor(n.App().factory, client.NodeGVR)
			if err != nil {
				n.App().Flash().Err(err)
				return
			}
			m, ok := res.(dao.NodeMaintainer)
			if !ok {
				n.App().Flash().Err(fmt.Errorf("expecting a maintainer for %q", client.NodeGVR))
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

func (n *NodeAlloc) sshCmd(evt *tcell.EventKey) *tcell.EventKey {
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

func (n *NodeAlloc) yamlCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := n.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	n.Stop()
	defer n.Start()
	ctx, cancel := context.WithTimeout(context.Background(), n.App().Conn().Config().CallTimeout())
	defer cancel()

	sel := n.GetTable().GetSelectedItem()
	gvr := client.NodeGVR.GVR()
	dial, err := n.App().factory.Client().DynDial()
	if err != nil {
		n.App().Flash().Err(err)
		return nil
	}
	o, err := dial.Resource(gvr).Get(ctx, sel, metav1.GetOptions{})
	if err != nil {
		n.App().Flash().Errf("Unable to get resource %q -- %s", client.NodeGVR, err)
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
