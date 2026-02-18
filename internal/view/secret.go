// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"os"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/labels"
)

// Secret presents a secret viewer.
type Secret struct {
	ResourceViewer
}

// NewSecret returns a new viewer.
func NewSecret(gvr *client.GVR) ResourceViewer {
	s := Secret{
		ResourceViewer: NewOwnerExtender(NewBrowser(gvr)),
	}
	s.AddBindKeysFn(s.bindKeys)

	return &s
}

func (s *Secret) bindKeys(aa *ui.KeyActions) {
	if !s.App().Config.IsReadOnly() {
		aa.Add(ui.KeyE, ui.NewKeyActionWithOpts("Edit", s.editCmd, ui.ActionOpts{
			Visible:   true,
			Dangerous: true,
		}))
	}

	aa.Bulk(ui.KeyMap{
		ui.KeyX: ui.NewKeyAction("Decode", s.decodeCmd, true),
		ui.KeyU: ui.NewKeyAction("UsedBy", s.refCmd, true),
	})
}

func (s *Secret) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	s.Stop()
	defer s.Start()
	if err := editSecretRes(s.App(), s.GVR(), path); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}

func (s *Secret) refCmd(evt *tcell.EventKey) *tcell.EventKey {
	return scanRefs(evt, s.App(), s.GetTable(), client.SecGVR)
}

func (s *Secret) decodeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	raw, err := decodedSecretYAML(s.App(), s.GVR(), path)
	if err != nil {
		s.App().Flash().Errf("Error decoding secret %s", err)
		return nil
	}

	details := NewDetails(s.App(), "Secret Decoder", path, contentYAML, true)
	details.SetEditFn(func() error {
		if err := editSecretRes(s.App(), s.GVR(), path); err != nil {
			return err
		}

		raw, err := decodedSecretYAML(s.App(), s.GVR(), path)
		if err != nil {
			return err
		}
		details.Update(raw)

		return nil
	}).
		Update(raw)
	if err := s.App().inject(details, false); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}

func decodedSecretYAML(app *App, gvr *client.GVR, path string) (string, error) {
	o, err := app.factory.Get(gvr, path, true, labels.Everything())
	if err != nil {
		return "", err
	}

	mm, err := dao.ExtractSecrets(o)
	if err != nil {
		return "", err
	}

	raw, err := data.WriteYAML(mm)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func editSecretRes(app *App, gvr *client.GVR, path string) error {
	if path == "" {
		return fmt.Errorf("nothing selected %q", path)
	}
	ns, n := client.Namespaced(path)
	if n == "" {
		return fmt.Errorf("missing resource name in path %q", path)
	}
	if ok, err := app.Conn().CanI(ns, gvr, n, client.PatchAccess); !ok || err != nil {
		return fmt.Errorf("current user can't edit resource %s", gvr)
	}

	res, err := dao.AccessorFor(app.factory, gvr)
	if err != nil {
		return fmt.Errorf("failed to get accessor: %w", err)
	}
	sec, ok := res.(*dao.Secret)
	if !ok {
		return fmt.Errorf("expecting Secret DAO but got %T", res)
	}

	yaml, err := sec.EditYAML(path)
	if err != nil {
		return fmt.Errorf("failed to get secret YAML: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "k9s-secret-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yaml); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if !edit(app, &shellOpts{clear: true, args: []string{tmpFile.Name()}}) {
		return fmt.Errorf("failed to launch editor")
	}

	args := make([]string, 0, 5)
	args = append(args, "replace", "-f", tmpFile.Name())
	if ns != client.BlankNamespace {
		args = append(args, "-n", ns)
	}

	return runK(app, &shellOpts{clear: false, args: args})
}
