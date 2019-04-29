package views

import (
	"sigs.k8s.io/yaml"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type secretView struct {
	*resourceView
}

func newSecretView(t string, app *appView, list resource.List) resourceViewer {
	v := secretView{newResourceView(t, app, list).(*resourceView)}
	{
		v.extraActionsFn = v.extraActions
		v.switchPage("secret")
	}

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
		v.app.flash(flashErr, "Unable to retrieve secret", sel)
		return evt
	}

	d := make(map[string]string, len(sec.Data))
	for k, val := range sec.Data {
		d[k] = string(val)
	}
	raw, err := yaml.Marshal(d)
	if err != nil {
		v.app.flash(flashErr, "Error decoding secret for `", sel)
		log.Error().Err(err).Msgf("Marshal error getting secret %s", sel)
		return nil
	}

	details := v.GetPrimitive("details").(*detailsView)
	{
		details.setCategory("Decoder")
		details.setTitle(sel)
		details.SetTextColor(tcell.ColorMediumAquamarine)
		details.SetText(colorizeYAML(v.app.styles.Style, string(raw)))
		details.ScrollToBeginning()
	}
	v.switchPage("details")

	return nil
}
