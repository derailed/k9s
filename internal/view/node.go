package view

import (
	"context"
	"fmt"
	"time"

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

	return &n
}

func (n *Node) bindDangerousKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyC: ui.NewKeyAction("Cordon", n.toggleCordonCmd(true), true),
		ui.KeyU: ui.NewKeyAction("Uncordon", n.toggleCordonCmd(false), true),
		ui.KeyR: ui.NewKeyAction("Drain", n.drainCmd, true),
	})
	cl := n.App().Config.K9s.CurrentCluster
	if n.App().Config.K9s.Clusters[cl].FeatureGates.NodeShell {
		aa.Add(ui.KeyActions{
			ui.KeyS: ui.NewKeyAction("Shell", n.sshCmd, true),
		})
	}
}

func (n *Node) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeySpace, tcell.KeyCtrlSpace)

	if !n.App().Config.K9s.IsReadOnly() {
		n.bindDangerousKeys(aa)
	}

	aa.Add(ui.KeyActions{
		ui.KeyY:      ui.NewKeyAction("YAML", n.yamlCmd, true),
		ui.KeyShiftC: ui.NewKeyAction("Sort CPU", n.GetTable().SortColCmd(cpuCol, false), false),
		ui.KeyShiftM: ui.NewKeyAction("Sort MEM", n.GetTable().SortColCmd(memCol, false), false),
		ui.KeyShift0: ui.NewKeyAction("Sort Pods", n.GetTable().SortColCmd("PODS", false), false),
	})
}

func (n *Node) showPods(a *App, _ ui.Tabular, _, path string) {
	showPods(a, n.GetTable().GetSelectedItem(), client.AllNamespaces, "spec.nodeName="+path)
}

func (n *Node) drainCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := n.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	opts := dao.DrainOptions{
		GracePeriodSeconds: -1,
		Timeout:            5 * time.Second,
	}
	ShowDrain(n, path, opts, drainNode)

	return nil
}

func drainNode(v ResourceViewer, path string, opts dao.DrainOptions) {
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
		d := NewDetails(v.App(), "Drain Progress", path, true)
		if err := v.App().inject(d, false); err != nil {
			v.App().Flash().Err(err)
		}

		if err := m.Drain(path, opts, d.GetWriter()); err != nil {
			v.App().Flash().Err(err)
			return
		}
		v.Refresh()
	}
}

func (n *Node) toggleCordonCmd(cordon bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		path := n.GetTable().GetSelectedItem()
		if path == "" {
			return evt
		}

		title, msg := "Confirm ", ""
		if cordon {
			title, msg = title+"Cordon", "Cordon "
		} else {
			title, msg = title+"Uncordon", "Uncordon "
		}
		msg += path + "?"
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
			if err := m.ToggleCordon(path, cordon); err != nil {
				n.App().Flash().Err(err)
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
	if err := ssh(n.App(), node); err != nil {
		log.Error().Err(err).Msgf("SSH Failed")
	}

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

	details := NewDetails(n.App(), "YAML", sel, true).Update(raw)
	if err := n.App().inject(details, false); err != nil {
		n.App().Flash().Err(err)
	}

	return nil
}
