package view

import (
	"context"

	"sigs.k8s.io/yaml"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Secret presents a secret viewer.
type Secret struct {
	ResourceViewer
}

// NewSecrets returns a new viewer.
func NewSecret(title, gvr string, list resource.List) ResourceViewer {
	return &Secret{
		ResourceViewer: NewResource(title, gvr, list),
	}
}

func (s *Secret) Init(ctx context.Context) error {
	if err := s.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	s.bindKeys()

	return nil
}

func (s *Secret) bindKeys() {
	s.Actions().Add(ui.KeyActions{
		tcell.KeyCtrlX: ui.NewKeyAction("Decode", s.decodeCmd, true),
	})
}

func (s *Secret) decodeCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := s.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	ns, n := namespaced(sel)
	sec, err := s.App().Conn().DialOrDie().CoreV1().Secrets(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		s.App().Flash().Errf("Unable to retrieve secret %s", err)
		return evt
	}

	d := make(map[string]string, len(sec.Data))
	for k, val := range sec.Data {
		d[k] = string(val)
	}
	raw, err := yaml.Marshal(d)
	if err != nil {
		s.App().Flash().Errf("Error decoding secret %s", err)
		return nil
	}

	details := NewDetails("Decoder")
	details.SetSubject(sel)
	details.SetTextColor(s.App().Styles.FgColor())
	details.SetText(colorizeYAML(s.App().Styles.Views().Yaml, string(raw)))
	details.ScrollToBeginning()
	s.App().inject(details)

	return nil
}
