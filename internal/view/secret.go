// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/slogs"
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
	aa.Bulk(ui.KeyMap{
		ui.KeyX: ui.NewKeyAction("Decode", s.decodeCmd, true),
		ui.KeyU: ui.NewKeyAction("UsedBy", s.refCmd, true),
		ui.KeyE: ui.NewKeyActionWithOpts("Edit Decoded", s.editDecodedCmd,
			ui.ActionOpts{Visible: true, Dangerous: true}),
		ui.KeyShiftE: ui.NewKeyActionWithOpts("Edit Raw", s.editRawCmd,
			ui.ActionOpts{Visible: true, Dangerous: true}),
	})
}

func (s *Secret) refCmd(evt *tcell.EventKey) *tcell.EventKey {
	return scanRefs(evt, s.App(), s.GetTable(), client.SecGVR)
}

func (s *Secret) editRawCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	s.Stop()
	defer s.Start()
	if err := editRes(s.App(), s.GVR(), path); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}

func (s *Secret) editDecodedCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	s.Stop()
	defer s.Start()
	if err := editDecodedSecret(s.App(), path); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}

func editDecodedSecret(app *App, path string) error {
	ns, n := client.Namespaced(path)
	if n == "" {
		return fmt.Errorf("missing resource name in path %q", path)
	}
	if client.IsClusterScoped(ns) {
		ns = client.BlankNamespace
	}

	ok, err := app.Conn().CanI(ns, client.SecGVR, n, client.PatchAccess)
	if !ok || err != nil {
		return fmt.Errorf("current user can't edit secret %s", path)
	}

	var sec dao.Secret
	sec.Init(app.factory, client.SecGVR)

	original, err := sec.GetEditableYAML(path)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "k9s-secret-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if e := os.Remove(tmpFile.Name()); e != nil {
			slog.Debug("Failed to remove temp file", slogs.Error, e)
		}
	}()

	if _, err := tmpFile.Write(original); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	if !edit(app, &shellOpts{clear: true, args: []string{tmpFile.Name()}}) {
		return fmt.Errorf("editor command failed")
	}

	edited, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	if bytes.Equal(original, edited) {
		app.Flash().Info("Edit cancelled, no changes made")
		return nil
	}

	if err := sec.UpdateFromEditedYAML(edited); err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	app.Flash().Infof("Secret %s updated successfully", path)

	return nil
}

func (s *Secret) decodeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	o, err := s.App().factory.Get(s.GVR(), path, true, labels.Everything())
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	mm, err := dao.ExtractSecrets(o)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}

	raw, err := data.WriteYAML(mm)
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
