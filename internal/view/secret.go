package view

import (
	"sigs.k8s.io/yaml"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Secret presents a secret viewer.
type Secret struct {
	*Resource
}

// NewSecrets returns a new viewer.
func NewSecret(title, gvr string, list resource.List) ResourceViewer {
	s := Secret{
		Resource: NewResource(title, gvr, list),
	}
	s.extraActionsFn = s.extraActions

	return &s
}

func (s *Secret) extraActions(aa ui.KeyActions) {
	aa[tcell.KeyCtrlX] = ui.NewKeyAction("Decode", s.decodeCmd, true)
}

func (s *Secret) decodeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.masterPage().RowSelected() {
		return evt
	}

	sel := s.masterPage().GetSelectedItem()
	ns, n := namespaced(sel)
	sec, err := s.app.Conn().DialOrDie().CoreV1().Secrets(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		s.app.Flash().Errf("Unable to retrieve secret %s", err)
		return evt
	}

	d := make(map[string]string, len(sec.Data))
	for k, val := range sec.Data {
		d[k] = string(val)
	}
	raw, err := yaml.Marshal(d)
	if err != nil {
		s.app.Flash().Errf("Error decoding secret %s", err)
		return nil
	}

	details := s.detailsPage()
	details.setCategory("Decoder")
	details.setTitle(sel)
	details.SetTextColor(s.app.Styles.FgColor())
	details.SetText(colorizeYAML(s.app.Styles.Views().Yaml, string(raw)))
	details.ScrollToBeginning()
	s.showDetails()

	return nil
}
