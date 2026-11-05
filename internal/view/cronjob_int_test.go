// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

type cronJobViewerStub struct {
	ResourceViewer

	app *App
}

func (s *cronJobViewerStub) App() *App {
	return s.app
}

func TestCronJobBindKeysReadOnly(t *testing.T) {
	uu := map[string]struct {
		readOnly bool
	}{
		"read-write": {readOnly: false},
		"read-only":  {readOnly: true},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			cfg := mock.NewMockConfig(t)
			cfg.K9s.ReadOnly = u.readOnly
			c := CronJob{ResourceViewer: &cronJobViewerStub{app: NewApp(cfg)}}

			aa := ui.NewKeyActions()
			c.bindKeys(aa)

			_, hasTrigger := aa.Get(ui.KeyT)
			_, hasSuspend := aa.Get(ui.KeyS)
			assert.Equal(t, !u.readOnly, hasTrigger)
			assert.Equal(t, !u.readOnly, hasSuspend)
		})
	}
}
