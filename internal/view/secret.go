package view

import (
	"sigs.k8s.io/yaml"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Secret presents a secret viewer.
type Secret struct {
	ResourceViewer
}

// NewSecret returns a new viewer.
func NewSecret(gvr client.GVR) ResourceViewer {
	s := Secret{
		ResourceViewer: NewBrowser(gvr),
	}
	s.SetBindKeysFn(s.bindKeys)

	return &s
}

func (s *Secret) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyX: ui.NewKeyAction("Decode", s.decodeCmd, true),
		ui.KeyR: ui.NewKeyAction("Referenced by", s.refCmd, true),
	})
}

func (s *Secret) refCmd(evt *tcell.EventKey) *tcell.EventKey {
	return scanRefs(evt, s.App(), s.GetTable(), "v1/secrets")
}

func (s *Secret) decodeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	o, err := s.App().factory.Get(s.GVR().String(), path, true, labels.Everything())
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	var secret v1.Secret
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &secret)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	d := make(map[string]string, len(secret.Data))
	for k, val := range secret.Data {
		d[k] = string(val)
	}
	raw, err := yaml.Marshal(d)
	if err != nil {
		s.App().Flash().Errf("Error decoding secret %s", err)
		return nil
	}

	details := NewDetails(s.App(), "Secret Decoder", path, true).Update(string(raw))
	if err := s.App().inject(details); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}
