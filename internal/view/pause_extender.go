// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/config"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
)

// PauseExtender adds pausing extensions.
type PauseExtender struct {
	ResourceViewer
}

// NewPauseExtender returns a new extender.
func NewPauseExtender(r ResourceViewer) ResourceViewer {
	p := PauseExtender{ResourceViewer: r}
	p.AddBindKeysFn(p.bindKeys)

	return &p
}

const (
	PAUSE        = "Pause"
	RESUME       = "Resume"
	PAUSE_RESUME = "Pause/Resume"
)

func (p *PauseExtender) bindKeys(aa *ui.KeyActions) {
	if p.App().Config.K9s.IsReadOnly() {
		return
	}

	aa.Add(ui.KeyZ, ui.NewKeyActionWithOpts(PAUSE_RESUME, p.togglePauseCmd,
		ui.ActionOpts{
			Visible:   true,
			Dangerous: true,
		},
	))
}

func (p *PauseExtender) togglePauseCmd(evt *tcell.EventKey) *tcell.EventKey {
	paths := p.GetTable().GetSelectedItems()
	if len(paths) == 0 {
		return nil
	}

	p.Stop()
	defer p.Start()

	styles := p.App().Styles.Dialog()
	form := p.makeStyledForm(styles)

	action := PAUSE
	if len(paths) == 1 {
		isPaused, err := p.valueOf("PAUSED")
		if err != nil {
			log.Error().Err(err).Msg("Reading 'PAUSED' state failed")
			p.App().Flash().Err(err)
			return nil
		}

		if isPaused == "true" {
			action = RESUME
		}
	}

	if len(paths) > 1 {
		form.AddDropDown("Action:", []string{PAUSE, RESUME}, 0, func(option string, optionIndex int) {
			action = option
		})
	}

	form.AddButton("OK", func() {
		defer p.dismissDialog()

		ctx, cancel := context.WithTimeout(context.Background(), p.App().Conn().Config().CallTimeout())
		defer cancel()

		for _, sel := range paths {
			if err := p.togglePause(ctx, sel, action); err != nil {
				log.Error().Err(err).Msgf("DP %s pausing failed", sel)
				p.App().Flash().Err(err)
				return
			}
		}

		if len(paths) == 1 {
			p.App().Flash().Infof("[%d] %s paused successfully", len(paths), singularize(p.GVR().R()))
		} else {
			p.App().Flash().Infof("%s %s paused successfully", p.GVR().R(), paths[0])
		}
	})

	form.AddButton("Cancel", func() {
		p.dismissDialog()
	})
	for i := 0; i < 2; i++ {
		if b := form.GetButton(i); b != nil {
			b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
			b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
		}
	}

	confirm := tview.NewModalForm("Pause/Resume", form)
	msg := fmt.Sprintf("%s %s %s?", action, singularize(p.GVR().R()), paths[0])
	if len(paths) > 1 {
		msg = fmt.Sprintf("Pause/Resume [%d] %s?", len(paths), p.GVR().R())
	}
	confirm.SetText(msg)
	confirm.SetDoneFunc(func(int, string) {
		p.dismissDialog()
	})
	p.App().Content.AddPage(pauseDialogKey, confirm, false, false)
	p.App().Content.ShowPage(pauseDialogKey)

	return nil
}

func (p *PauseExtender) togglePause(ctx context.Context, path string, action string) error {
	res, err := dao.AccessorFor(p.App().factory, p.GVR())
	if err != nil {
		p.App().Flash().Err(err)
		return nil
	}
	pauser, ok := res.(dao.Pausable)
	if !ok {
		p.App().Flash().Err(fmt.Errorf("expecting a pausable resource for %q", p.GVR()))
		return nil
	}

	if action == PAUSE {
		err = pauser.Pause(ctx, path)
	} else if action == RESUME {
		err = pauser.Resume(ctx, path)
	} else {
		p.App().Flash().Err(fmt.Errorf("failed to identify action; must be '%s' or '%s' but is: '%s'", PAUSE, RESUME, action))
		return nil
	}

	if err != nil {
		p.App().Flash().Err(fmt.Errorf("failed to %s: %q", action, err))
	}

	return nil
}

func (p *PauseExtender) valueOf(col string) (string, error) {
	colIdx, ok := p.GetTable().HeaderIndex(col)
	if !ok {
		return "", fmt.Errorf("no column index for %s", col)
	}
	return p.GetTable().GetSelectedCell(colIdx), nil
}

const pauseDialogKey = "pause"

func (p *PauseExtender) dismissDialog() {
	p.App().Content.RemovePage(pauseDialogKey)
}

func (p *PauseExtender) makeStyledForm(styles config.Dialog) *tview.Form {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonBgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())

	return f
}
