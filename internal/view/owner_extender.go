package view

import (
	"errors"
	"fmt"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const selectOwnerDialogKey = "owner"

// OwnerExtender adds log actions to a given viewer.
type OwnerExtender struct {
	ResourceViewer
}

// NewOwnerExtender returns a new extender.
func NewOwnerExtender(v ResourceViewer) ResourceViewer {
	o := OwnerExtender{
		ResourceViewer: v,
	}
	o.AddBindKeysFn(o.bindKeys)

	return &o
}

// BindKeys injects new menu actions.
func (o *OwnerExtender) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyO: ui.NewKeyAction("Show Owner", o.ownerCmd(), true),
	})
}

func (o *OwnerExtender) ownerCmd() func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		path := o.GetTable().GetSelectedItem()
		if path == "" {
			return nil
		}
		if !isResourcePath(path) {
			path = o.GetTable().Path
		}

		err := o.showOwner(path)
		if err != nil {
			o.App().Flash().Err(err)
		}

		return nil
	}
}

func (o *OwnerExtender) showOwner(path string) error {
	res, err := dao.AccessorFor(o.App().factory, o.GVR())
	if err != nil {
		return err
	}
	owned, ok := res.(dao.Owned)
	if !ok {
		return fmt.Errorf("owner navigation not possible for resource %q", o.GVR())
	}

	owners, err := owned.GetOwners(path)
	if err != nil {
		return err
	}

	if len(owners) == 0 {
		return errors.New("resource does not have an owner")
	}

	if len(owners) == 1 {
		o.goToOwner(owners[0])
		return nil
	}

	return o.showSelectOwnerDialog(owners)
}

func (o *OwnerExtender) goToOwner(owner dao.OwnerInfo) {
	o.App().gotoResource(owner.GVR.String(), owner.FQN, false)
}

func (o *OwnerExtender) showSelectOwnerDialog(refs []dao.OwnerInfo) error {
	form, err := o.makeSelectOwnerForm(refs)
	if err != nil {
		return err
	}
	modal := tview.NewModalForm("<Owner>", form)
	msg := "Select owner"
	modal.SetText(msg)
	modal.SetDoneFunc(func(int, string) {
		o.dismissDialog()
	})
	o.App().Content.AddPage(selectOwnerDialogKey, modal, false, false)
	o.App().Content.ShowPage(selectOwnerDialogKey)

	return nil
}

func (o *OwnerExtender) makeSelectOwnerForm(refs []dao.OwnerInfo) (*tview.Form, error) {
	f := o.makeStyledForm()

	var ownerLabels []string
	for _, ref := range refs {
		ownerLabels = append(ownerLabels, fmt.Sprintf("<%s> %s", ref.GVR, ref.FQN))
	}

	var selectedRef dao.OwnerInfo

	f.AddDropDown("Owner:", ownerLabels, 0, func(option string, optionIndex int) {
		selectedRef = refs[optionIndex]
		return
	})

	f.AddButton("OK", func() {
		defer o.dismissDialog()
		o.goToOwner(selectedRef)
	})

	f.AddButton("Cancel", func() {
		o.dismissDialog()
	})

	return f, nil
}

func (o *OwnerExtender) makeStyledForm() *tview.Form {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)

	return f
}

func (o *OwnerExtender) dismissDialog() {
	o.App().Content.RemovePage(selectOwnerDialogKey)
}
