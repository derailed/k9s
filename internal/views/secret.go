package views

import (
	"sigs.k8s.io/yaml"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type secretView struct {
	*resourceView
}

func newSecretView(t string, app *appView, list resource.List) resourceViewer {
	v := secretView{newResourceView(t, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions

	return &v
}

func (v *secretView) extraActions(aa keyActions) {
	aa[tcell.KeyCtrlX] = newKeyAction("Decode", v.decodeCmd, true)
}

func (v *secretView) decodeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	sel := v.getSelectedItem()
	ns, n := namespaced(sel)
	sec, err := v.app.conn().DialOrDie().CoreV1().Secrets(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		v.app.flash().errf("Unable to retrieve secret %s", err)
		return evt
	}

	d := make(map[string]string, len(sec.Data))
	for k, val := range sec.Data {
		d[k] = string(val)
	}
	raw, err := yaml.Marshal(d)
	if err != nil {
		v.app.flash().errf("Error decoding secret %s", err)
		return nil
	}

	details := v.detailsPage()
	details.setCategory("Decoder")
	details.setTitle(sel)
	details.SetTextColor(v.app.styles.FgColor())
	details.SetText(colorizeYAML(v.app.styles.Style, string(raw)))
	details.ScrollToBeginning()
	v.switchPage("details")

	return nil
}
