// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"
)

// Secret presents a secret viewer.
type Secret struct {
	ResourceViewer
}

// NewSecret returns a new viewer.
func NewSecret(gvr client.GVR) ResourceViewer {
	s := Secret{
		ResourceViewer: NewOwnerExtender(NewBrowser(gvr)),
	}
	s.AddBindKeysFn(s.bindKeys)

	return &s
}

func (s *Secret) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyX: ui.NewKeyAction("Decode", s.decodeCmd, true),
		ui.KeyU: ui.NewKeyAction("UsedBy", s.refCmd, true),
	})
}

func (s *Secret) refCmd(evt *tcell.EventKey) *tcell.EventKey {
	return scanRefs(evt, s.App(), s.GetTable(), dao.SecGVR)
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

	d, err := dao.ExtractSecrets(o.(*unstructured.Unstructured))
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	raw, err := yaml.Marshal(d)
	if err != nil {
		s.App().Flash().Errf("Error decoding secret %s", err)
		return nil
	}

	details := NewDetails(s.App(), "Secret Decoder", path, contentYAML, true).Update(string(raw))
	if err := s.App().inject(details, false); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}
