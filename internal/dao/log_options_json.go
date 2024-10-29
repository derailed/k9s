// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"
	"github.com/derailed/k9s/internal/config"
	"slices"
	"strings"
)

// JsonTemplateListener represents a json template selection listener.
type JsonTemplateListener interface {
	// JsonTemplateChanged indicates template was changed.
	JsonTemplateChanged()
}

type JsonTemplate struct {
	Name               string
	LogLevelExpression string
	DateTimeExpression string
	MessageExpression  string
}

// Clone clones options.
func (o *JsonTemplate) Clone() *JsonTemplate {
	return &JsonTemplate{
		Name:               o.Name,
		LogLevelExpression: o.LogLevelExpression,
		DateTimeExpression: o.DateTimeExpression,
		MessageExpression:  o.MessageExpression,
	}
}

type JsonOptions struct {
	GlobalExpressions    string
	CurrentTemplateIndex int
	Templates            []JsonTemplate
	listeners            []JsonTemplateListener
}

func TemplatesFromConfig(config config.JsonConfig) []JsonTemplate {
	var templates []JsonTemplate
	for _, obj := range config.Templates {
		templates = append(templates, JsonTemplate{
			Name:               obj.Name,
			LogLevelExpression: obj.LogLevelExpression,
			DateTimeExpression: obj.DateTimeExpression,
			MessageExpression:  obj.MessageExpression,
		})
	}
	if templates == nil {
		templates = []JsonTemplate{
			{
				Name:               "default",
				LogLevelExpression: ".level",
				DateTimeExpression: ".timestamp",
				MessageExpression:  ".message",
			},
		}
	}
	return templates
}

// SetCurrentTemplateByName Set the currently selected template.
func (o *JsonOptions) SetCurrentTemplateByName(templateName string) {
	o.CurrentTemplateIndex = 0
	o.CurrentTemplateIndex = slices.IndexFunc(o.Templates, func(j JsonTemplate) bool {
		return j.Name == templateName
	})
	if (o.CurrentTemplateIndex < 0) || (o.CurrentTemplateIndex >= len(o.Templates)) {
		o.CurrentTemplateIndex = 0
	}
	o.NotifyListeners()
}

// IterateToNextTemplate Iterates the currently selected template to next (back to first on end).
func (o *JsonOptions) IterateToNextTemplate() {
	o.CurrentTemplateIndex = (o.CurrentTemplateIndex + 1) % len(o.Templates)
	o.NotifyListeners()
}

// SetCurrentTemplate Set the currently selected template.
func (o *JsonOptions) SetCurrentTemplate(index int) *JsonTemplate {
	o.CurrentTemplateIndex = index
	o.NotifyListeners()
	return o.GetCurrentTemplate()
}

// UpdateCurrentTemplate Updates the currently selected template and notifies listeners.
func (o *JsonOptions) UpdateCurrentTemplate(LogLevelExpression string, DateTimeExpression string, MessageExpression string) *JsonTemplate {
	var template = o.GetCurrentTemplate()
	template.LogLevelExpression = LogLevelExpression
	template.DateTimeExpression = DateTimeExpression
	template.MessageExpression = MessageExpression
	o.NotifyListeners()
	return template
}

// GetCurrentTemplate Return the currently selected template.
func (o *JsonOptions) GetCurrentTemplate() *JsonTemplate {
	return &o.Templates[o.CurrentTemplateIndex]
}

// GetAllTemplateNames Return all template names.
func (o *JsonOptions) GetAllTemplateNames() []string {
	var names []string
	for _, obj := range o.Templates {
		names = append(names, obj.Name)
	}
	return names
}

func (o *JsonOptions) GetCurrentJsonQuery() string {
	var template = o.GetCurrentTemplate()
	var query = `%s . as $line | try ( capture("^(?<ts>[0-9-:]{8,10}[^0-9][0-9\\.:]{0,10}[^ ]+) (?<js>.*)") | .ts as $k8sts | .js | fromjson | { datetime:(%s), lvl:(%s), msg:(%s) } | "\(.datetime) \(.lvl) \(.msg)" ) catch $line`
	var globalExpression = strings.Trim(strings.Trim(strings.ReplaceAll(o.GlobalExpressions, "\n", " "), " "), ";")
	if len(globalExpression) > 0 {
		globalExpression += ";"
	}
	return fmt.Sprintf(query, globalExpression, template.DateTimeExpression, template.LogLevelExpression, template.MessageExpression)
}

// Clone clones options.
func (o *JsonOptions) Clone() *JsonOptions {
	cTemplates := make([]JsonTemplate, len(o.Templates))
	for i, t := range o.Templates {
		cTemplates[i] = *t.Clone()
	}
	return &JsonOptions{
		GlobalExpressions:    o.GlobalExpressions,
		CurrentTemplateIndex: o.CurrentTemplateIndex,
		Templates:            cTemplates,
	}
}

// NotifyListeners notifies all template listeners about a change.
func (o *JsonOptions) NotifyListeners() {
	for _, lis := range o.listeners {
		lis.JsonTemplateChanged()
	}
}

// AddListener adds a new model listener.
func (o *JsonOptions) AddListener(listener JsonTemplateListener) {
	o.listeners = append(o.listeners, listener)
}

// RemoveListener delete a listener from the list.
func (o *JsonOptions) RemoveListener(listener JsonTemplateListener) {
	victim := -1
	for i, lis := range o.listeners {
		if lis == listener {
			victim = i
			break
		}
	}

	if victim >= 0 {
		o.listeners = append(o.listeners[:victim], o.listeners[victim+1:]...)
	}
}
