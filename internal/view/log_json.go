// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// LogTemplateForm represents a json log template viewer.
type LogTemplateForm struct {
	model *dao.JsonOptions
	app   *App
	form  *tview.Form
	modal *tview.ModalForm
}

// NewLogTemplateForm returns a new json template form.
func NewLogTemplateForm(app *App, opts *dao.JsonOptions) *LogTemplateForm {
	l := LogTemplateForm{
		model: opts,
		app:   app,
	}

	return &l
}

const jsonTemplateDialogKey = "jsonTemplateDialog"
const labelLogLevel = "Log Level :"
const labelDateTime = "Date Time :"
const labelMessage = "Message   :"

func (l *LogTemplateForm) showJsonTemplatesCmd(_ *tcell.EventKey) *tcell.EventKey {
	currentTemplate := l.model.GetCurrentTemplate()

	form := tview.NewForm().
		AddDropDown("Template", l.model.GetAllTemplateNames(), l.model.CurrentTemplateIndex, l.jsonTemplateSelected).
		AddInputField(labelLogLevel, currentTemplate.LogLevelExpression, 0, nil, l.validateLogLevelExpression).
		AddInputField(labelDateTime, currentTemplate.DateTimeExpression, 0, nil, l.validateDateTimeExpression).
		AddInputField(labelMessage, currentTemplate.MessageExpression, 0, nil, l.validateMessageExpression).
		AddButton("Apply", l.applyNewJsonExpressions).
		AddButton("Quit", l.dismissDialog)

	form.SetItemPadding(0)

	styles := l.app.Styles.Dialog()
	form.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color()).
		SetFieldBackgroundColor(tcell.GetColor("darkslategray").TrueColor())

	modal := tview.NewModalForm(" JSON Expressions ", form)
	modal.SetText("Set field expressions")
	modal.SetDoneFunc(func(int, string) {
		l.dismissDialog()
	})
	l.form = form
	l.modal = modal
	l.model.AddListener(l)

	l.app.Content.AddPage(jsonTemplateDialogKey, modal, true, false)
	l.app.Content.ShowPage(jsonTemplateDialogKey)

	return nil
}

func (l *LogTemplateForm) JsonTemplateChanged() {
	var template = l.model.GetCurrentTemplate()
	if l.form != nil {
		l.form.GetFormItemByLabel(labelLogLevel).(*tview.InputField).SetText(template.LogLevelExpression)
		l.form.GetFormItemByLabel(labelDateTime).(*tview.InputField).SetText(template.DateTimeExpression)
		l.form.GetFormItemByLabel(labelMessage).(*tview.InputField).SetText(template.MessageExpression)
	}
}

func (l *LogTemplateForm) jsonTemplateSelected(_ string, optionIndex int) {
	if optionIndex == l.model.CurrentTemplateIndex {
		return
	}
	l.model.SetCurrentTemplate(optionIndex)
}

func (l *LogTemplateForm) applyNewJsonExpressions() {
	logLevelExpression := l.form.GetFormItemByLabel(labelLogLevel).(*tview.InputField).GetText()
	dateTimeExpression := l.form.GetFormItemByLabel(labelDateTime).(*tview.InputField).GetText()
	messageExpression := l.form.GetFormItemByLabel(labelMessage).(*tview.InputField).GetText()

	err := l.model.TestJsonQueryCode(logLevelExpression, dateTimeExpression, messageExpression)
	if err == nil {
		l.model.UpdateCurrentTemplate(logLevelExpression, dateTimeExpression, messageExpression)
		l.dismissDialog()
	}
}

func (l *LogTemplateForm) validateLogLevelExpression(logLevelExpression string) {
	dateTimeExpression := l.form.GetFormItemByLabel(labelDateTime).(*tview.InputField).GetText()
	messageExpression := l.form.GetFormItemByLabel(labelMessage).(*tview.InputField).GetText()
	_ = l.validateExpressions(logLevelExpression, dateTimeExpression, messageExpression)
}

func (l *LogTemplateForm) validateDateTimeExpression(dateTimeExpression string) {
	logLevelExpression := l.form.GetFormItemByLabel(labelLogLevel).(*tview.InputField).GetText()
	messageExpression := l.form.GetFormItemByLabel(labelMessage).(*tview.InputField).GetText()
	_ = l.validateExpressions(logLevelExpression, dateTimeExpression, messageExpression)
}

func (l *LogTemplateForm) validateMessageExpression(messageExpression string) {
	logLevelExpression := l.form.GetFormItemByLabel(labelLogLevel).(*tview.InputField).GetText()
	dateTimeExpression := l.form.GetFormItemByLabel(labelDateTime).(*tview.InputField).GetText()
	_ = l.validateExpressions(logLevelExpression, dateTimeExpression, messageExpression)
}

func (l *LogTemplateForm) validateExpressions(logLevelExpression string, dateTimeExpression string, messageExpression string) error {
	err := l.model.TestJsonQueryCode(logLevelExpression, dateTimeExpression, messageExpression)
	if err != nil {
		l.modal.SetTextColor(l.app.Styles.Frame().Status.ErrorColor.Color())
		l.modal.SetText(fmt.Sprintf("Set field expressions\nProblem: %s", err.Error()))
	} else {
		l.modal.SetTextColor(l.app.Styles.Dialog().FgColor.Color())
		l.modal.SetText("Set field expressions")
	}
	return err
}

func (l *LogTemplateForm) dismissDialog() {
	l.form = nil
	l.model.RemoveListener(l)
	l.app.Content.RemovePage(jsonTemplateDialogKey)
}
