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
	i := ui.NewStatusIndicator(ui.NewApp(mockConfig, ""), config.NewStyles())

	mockClustMetaCurr := mock.NewMockClusterMeta()
	mockClustMetaCurr.Cpu += 1
	mockClustMetaCurr.Mem += 1
	mockClustMetaPrev := mock.NewMockClusterMeta()

	actualStatText := i.BuildStatusIndicatorText(nil, &mockClustMetaCurr)
	actualStatTextWithPrev := i.BuildStatusIndicatorText(&mockClustMetaPrev, &mockClustMetaCurr)
	expectedStatText := DefaultStatusIndicatorText(nil, &mockClustMetaCurr)
	expectedStatTextWithPrev := DefaultStatusIndicatorText(&mockClustMetaPrev, &mockClustMetaCurr)

	assert.Equal(t, expectedStatText, actualStatText)
	assert.Equal(t, expectedStatTextWithPrev, actualStatTextWithPrev)
}

// If %s count does not match the given fields then use default
func TestIndicatorUseDefaultTextOnInvalidFMTConf(t *testing.T) {
	mockConfig := mock.NewMockConfig()
	mockConfig.K9s.UI.StatusIndicator.Format = "[orange::b]K9s [aqua::]%s [white::]%s"
	i := ui.NewStatusIndicator(ui.NewApp(mockConfig, ""), config.NewStyles())

	mockClustMetaCurr := mock.NewMockClusterMeta()
	mockClustMetaCurr.Cpu += 1
	mockClustMetaCurr.Mem += 1
	mockClustMetaPrev := mock.NewMockClusterMeta()

	actualStatText := i.BuildStatusIndicatorText(nil, &mockClustMetaCurr)
	actualStatTextWithPrev := i.BuildStatusIndicatorText(&mockClustMetaPrev, &mockClustMetaCurr)
	expectedStatText := DefaultStatusIndicatorText(nil, &mockClustMetaCurr)
	expectedStatTextWithPrev := DefaultStatusIndicatorText(&mockClustMetaPrev, &mockClustMetaCurr)

	assert.Equal(t, expectedStatText, actualStatText)
	assert.Equal(t, expectedStatTextWithPrev, actualStatTextWithPrev)
}

// If %s count does not match the given fields then use default
func TestIndicatorUseDefaultTextOnInvalidFieldsConf(t *testing.T) {
	mockConfig := mock.NewMockConfig()
	mockConfig.K9s.UI.StatusIndicator.Fields = []string{"CONTEXT"}
	i := ui.NewStatusIndicator(ui.NewApp(mockConfig, ""), config.NewStyles())

	mockClustMetaCurr := mock.NewMockClusterMeta()
	mockClustMetaCurr.Cpu += 1
	mockClustMetaCurr.Mem += 1
	mockClustMetaPrev := mock.NewMockClusterMeta()

	actualStatText := i.BuildStatusIndicatorText(nil, &mockClustMetaCurr)
	actualStatTextWithPrev := i.BuildStatusIndicatorText(&mockClustMetaPrev, &mockClustMetaCurr)
	expectedStatText := DefaultStatusIndicatorText(nil, &mockClustMetaCurr)
	expectedStatTextWithPrev := DefaultStatusIndicatorText(&mockClustMetaPrev, &mockClustMetaCurr)

	assert.Equal(t, expectedStatText, actualStatText)
	assert.Equal(t, expectedStatTextWithPrev, actualStatTextWithPrev)
}

// Check with valid config and fields if the indicator text is formatted correctly
func TestIndicatorTextFromConfig(t *testing.T) {
	mockConfig := mock.NewMockConfig()
	mockConfig.K9s.UI.StatusIndicator.Format = "[orange::b]MYTEXT %s"
	mockConfig.K9s.UI.StatusIndicator.Fields = []string{"CONTEXT"}
	i := ui.NewStatusIndicator(ui.NewApp(mockConfig, ""), config.NewStyles())

	mockClustMetaCurr := mock.NewMockClusterMeta()
	mockClustMetaCurr.Cpu += 1
	mockClustMetaCurr.Mem += 1
	mockClustMetaPrev := mock.NewMockClusterMeta()

	actualStatText := i.BuildStatusIndicatorText(nil, &mockClustMetaCurr)
	actualStatTextWithPrev := i.BuildStatusIndicatorText(&mockClustMetaPrev, &mockClustMetaCurr)
	expectedStatText := fmt.Sprintf("[orange::b]MYTEXT %s", mockClustMetaCurr.Context)

	assert.Equal(t, expectedStatText, actualStatText)
	assert.Equal(t, expectedStatText, actualStatTextWithPrev)
}

// Helper funcs
func DefaultStatusIndicatorText(prev, curr *model.ClusterMeta) string {
	var (
		cpuPerc string
		memPerc string
	)
	if prev != nil {
		cpuPerc = ui.AsPercDelta(prev.Cpu, curr.Cpu)
		memPerc = ui.AsPercDelta(prev.Cpu, curr.Mem)
	} else {
		cpuPerc = render.PrintPerc(curr.Cpu)
		memPerc = render.PrintPerc(curr.Mem)
	}

	defaultformat := "[orange::b]K9s [aqua::]%s [white::]%s:%s:%s [lawngreen::]%s[white::]::[darkturquoise::]%s"
	return fmt.Sprintf(
		defaultformat,
		curr.K9sVer,
		curr.Context,
		curr.Cluster,
		curr.K8sVer,
		cpuPerc,
		memPerc)
}
