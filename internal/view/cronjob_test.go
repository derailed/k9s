// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestContainerSectionTitle(t *testing.T) {
	assert.Equal(t, "[migrate] [init]", containerSectionTitle("migrate", true))
	assert.Equal(t, "[worker]", containerSectionTitle("worker", false))
}

func TestTriggerSectionPreservesNavigationDirection(t *testing.T) {
	navigation := tcell.KeyTab
	section := newTriggerSection("[worker]", &navigation, tcell.ColorBlue)
	var got tcell.Key
	section.SetFinishedFunc(func(key tcell.Key) { got = key })

	section.Focus(nil)
	assert.Equal(t, tcell.KeyTab, got)

	navigation = tcell.KeyBacktab
	section.Focus(nil)
	assert.Equal(t, tcell.KeyBacktab, got)
}

func TestStringSliceJSON(t *testing.T) {
	assert.Equal(t, "[]", stringSliceJSON(nil))
	assert.Equal(t, `["sh","-c"]`, stringSliceJSON([]string{"sh", "-c"}))
}

func TestEnvSliceJSON(t *testing.T) {
	assert.Equal(t, "[]", envSliceJSON(nil))
	assert.Equal(t, `[{"name":"FOO","value":"bar"}]`, envSliceJSON([]v1.EnvVar{{Name: "FOO", Value: "bar"}}))
}

func TestTriggerOverrides(t *testing.T) {
	fields := []triggerContainerFields{{
		name:    "main",
		command: tview.NewInputField().SetText(`["sh","-c"]`),
		args:    tview.NewInputField().SetText(`["echo hello"]`),
		env:     tview.NewInputField().SetText(`[{"name":"FOO","value":"bar"}]`),
	}}

	overrides, err := triggerOverrides(fields)
	require.NoError(t, err)
	assert.Equal(t, []string{"sh", "-c"}, overrides["main"].Command)
	assert.Equal(t, []string{"echo hello"}, overrides["main"].Args)
	assert.Equal(t, []v1.EnvVar{{Name: "FOO", Value: "bar"}}, overrides["main"].Env)
}

func TestTriggerOverridesRejectsInvalidJSON(t *testing.T) {
	fields := []triggerContainerFields{{
		name:    "main",
		command: tview.NewInputField().SetText("sh -c"),
		args:    tview.NewInputField().SetText("[]"),
		env:     tview.NewInputField().SetText("[]"),
	}}

	_, err := triggerOverrides(fields)
	require.EqualError(t, err, `invalid command for container "main"`)
}

func TestTriggerOverridesRejectsInvalidEnvVariables(t *testing.T) {
	fields := []triggerContainerFields{{
		name:    "main",
		command: tview.NewInputField().SetText("[]"),
		args:    tview.NewInputField().SetText("[]"),
		env:     tview.NewInputField().SetText(`{"FOO":"bar"}`),
	}}

	_, err := triggerOverrides(fields)
	require.EqualError(t, err, `invalid env variables for container "main"`)
}
