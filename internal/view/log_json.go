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
		AddDropDown("Template  :",
			l.model.GetAllTemplateNames(),
			l.model.CurrentTemplateIndex,
			l.jsonTemplateSelected).
		AddInputField(labelLogLevel, currentTemplate.LogLevelExpression, 0, nil, nil).
		AddInputField(labelDateTime, currentTemplate.DateTimeExpression, 0, nil, nil).
		AddInputField(labelMessage, currentTemplate.MessageExpression, 0, nil, nil).
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

	confirm := tview.NewModalForm(" JSON Expressions ", form)
	confirm.SetText(fmt.Sprintf("Set field expressions"))
	confirm.SetDoneFunc(func(int, string) {
		l.dismissDialog()
	})
	l.form = form
	l.model.AddListener(l)

	l.app.Content.AddPage(jsonTemplateDialogKey, confirm, true, false)
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
	l.model.UpdateCurrentTemplate(logLevelExpression, dateTimeExpression, messageExpression)
	l.dismissDialog()
}

func (l *LogTemplateForm) dismissDialog() {
	l.form = nil
	l.model.RemoveListener(l)
	l.app.Content.RemovePage(jsonTemplateDialogKey)
}
