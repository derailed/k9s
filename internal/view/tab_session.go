// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import "github.com/derailed/k9s/internal/model"

// TabSession holds the isolated state for a single browser tab.
// Multiple sessions share the same factory, clusterModel and ui.App.
type TabSession struct {
	id            int
	Content       *PageStack
	command       *Command
	cmdHistory    *model.History
	filterHistory *model.History
	label         string
}
