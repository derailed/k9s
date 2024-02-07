// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestIndicatorReset(t *testing.T) {
	i := ui.NewStatusIndicator(ui.NewApp(mock.NewMockConfig(), ""), config.NewStyles())
	i.SetPermanent("Blee")
	i.Info("duh")
	i.Reset()

	assert.Equal(t, "Blee\n", i.GetText(false))
}

func TestIndicatorInfo(t *testing.T) {
	i := ui.NewStatusIndicator(ui.NewApp(mock.NewMockConfig(), ""), config.NewStyles())
	i.Info("Blee")

	assert.Equal(t, "[lawngreen::b] <Blee> \n", i.GetText(false))
}

func TestIndicatorWarn(t *testing.T) {
	i := ui.NewStatusIndicator(ui.NewApp(mock.NewMockConfig(), ""), config.NewStyles())
	i.Warn("Blee")

	assert.Equal(t, "[mediumvioletred::b] <Blee> \n", i.GetText(false))
}

func TestIndicatorErr(t *testing.T) {
	i := ui.NewStatusIndicator(ui.NewApp(mock.NewMockConfig(), ""), config.NewStyles())
	i.Err("Blee")

	assert.Equal(t, "[orangered::b] <Blee> \n", i.GetText(false))
}

// Checks the default status indicator text
func TestIndicatorDefaultText(t *testing.T) {
	mockConfig := mock.NewMockConfig()
	mockClustMeta := mock.NewMockClusterMeta()
	i := ui.NewStatusIndicator(ui.NewApp(mockConfig, ""), config.NewStyles())

	actualStatText := i.BuildStatusIndicatorText(nil, &mockClustMeta)
	expectedStatText := DefaultStatusIndicatorText(mockClustMeta)

	assert.Equal(t, expectedStatText, actualStatText)
}

// If %s count does not match the given fields then use default
func TestIndicatorUseDefaultTextOnInvalidFMTConf(t *testing.T) {
	mockConfig := mock.NewMockConfig()
	mockConfig.K9s.UI.StatusIndicator.Format = "[orange::b]K9s [aqua::]%s [white::]%s"
	mockClustMeta := mock.NewMockClusterMeta()
	i := ui.NewStatusIndicator(ui.NewApp(mockConfig, ""), config.NewStyles())

	actualStatText := i.BuildStatusIndicatorText(nil, &mockClustMeta)
	expectedStatText := DefaultStatusIndicatorText(mockClustMeta)

	assert.Equal(t, expectedStatText, actualStatText)
}

// If %s count does not match the given fields then use default
func TestIndicatorUseDefaultTextOnInvalidFieldsConf(t *testing.T) {
	mockConfig := mock.NewMockConfig()
	mockConfig.K9s.UI.StatusIndicator.Fields = []string{"CONTEXT"}
	mockClustMeta := mock.NewMockClusterMeta()
	i := ui.NewStatusIndicator(ui.NewApp(mockConfig, ""), config.NewStyles())

	actualStatText := i.BuildStatusIndicatorText(nil, &mockClustMeta)
	expectedStatText := DefaultStatusIndicatorText(mockClustMeta)

	assert.Equal(t, expectedStatText, actualStatText)
}

// Check with valid config and fields if the indicator text is formatted correctly
func TestIndicatorTextFromConfig(t *testing.T) {
	mockConfig := mock.NewMockConfig()
	mockClustMeta := mock.NewMockClusterMeta()
	i := ui.NewStatusIndicator(ui.NewApp(mockConfig, ""), config.NewStyles())

	mockConfig.K9s.UI.StatusIndicator.Format = "[orange::b]MYTEXT %s"
	mockConfig.K9s.UI.StatusIndicator.Fields = []string{"CONTEXT"}
	actualStatText := i.BuildStatusIndicatorText(nil, &mockClustMeta)

	expectedStatText := fmt.Sprintf("[orange::b]MYTEXT %s", mockClustMeta.Context)
	assert.Equal(t, expectedStatText, actualStatText)
}

// Helper funcs
func DefaultStatusIndicatorText(meta model.ClusterMeta) string {
	defaultformat := "[orange::b]K9s [aqua::]%s [white::]%s:%s:%s [lawngreen::]%s[white::]::[darkturquoise::]%s"

	return fmt.Sprintf(
		defaultformat,
		meta.K9sVer,
		meta.Context,
		meta.Cluster,
		meta.K8sVer,
		render.PrintPerc(meta.Cpu),
		render.PrintPerc(meta.Mem))
}
