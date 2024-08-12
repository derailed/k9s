// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_k9sOverrides(t *testing.T) {
	var (
		true = true
		cmd  = "po"
		dir  = "/tmp/blee"
	)

	uu := map[string]struct {
		k              *K9s
		rate           int
		ro, hl, cl, ll bool
	}{
		"plain": {
			k: &K9s{
				LiveViewAutoRefresh: false,
				ScreenDumpDir:       "",
				RefreshRate:         10,
				MaxConnRetry:        0,
				ReadOnly:            false,
				NoExitOnCtrlC:       false,
				UI:                  UI{},
				SkipLatestRevCheck:  false,
				DisablePodCounting:  false,
			},
			rate: 10,
		},
		"set": {
			k: &K9s{
				LiveViewAutoRefresh: false,
				ScreenDumpDir:       "",
				RefreshRate:         10,
				MaxConnRetry:        0,
				ReadOnly:            true,
				NoExitOnCtrlC:       false,
				UI: UI{
					Headless:   true,
					Logoless:   true,
					Crumbsless: true,
				},
				SkipLatestRevCheck: false,
				DisablePodCounting: false,
			},
			rate: 10,
			ro:   true,
			hl:   true,
			ll:   true,
			cl:   true,
		},
		"overrides": {
			k: &K9s{
				LiveViewAutoRefresh: false,
				ScreenDumpDir:       "",
				RefreshRate:         10,
				MaxConnRetry:        0,
				ReadOnly:            false,
				NoExitOnCtrlC:       false,
				UI: UI{
					Headless:   false,
					Logoless:   false,
					Crumbsless: false,
				},
				SkipLatestRevCheck:  false,
				DisablePodCounting:  false,
				manualRefreshRate:   100,
				manualReadOnly:      &true,
				manualHeadless:      &true,
				manualLogoless:      &true,
				manualCrumbsless:    &true,
				manualCommand:       &cmd,
				manualScreenDumpDir: &dir,
			},
			rate: 100,
			ro:   true,
			hl:   true,
			ll:   true,
			cl:   true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.rate, u.k.GetRefreshRate())
			assert.Equal(t, u.ro, u.k.IsReadOnly())
			assert.Equal(t, u.cl, u.k.IsCrumbsless())
			assert.Equal(t, u.hl, u.k.IsHeadless())
			assert.Equal(t, u.ll, u.k.IsLogoless())

		})
	}
}

func Test_screenDumpDirOverride(t *testing.T) {
	uu := map[string]struct {
		dir string
		e   string
	}{
		"empty": {
			e: "/tmp/k9s-test/screen-dumps",
		},
		"override": {
			dir: "/tmp/k9s-test/sd",
			e:   "/tmp/k9s-test/sd",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cfg := NewConfig(nil)
			assert.NoError(t, cfg.Load("testdata/configs/k9s.yaml", true))

			cfg.K9s.manualScreenDumpDir = &u.dir
			assert.Equal(t, u.e, cfg.K9s.AppScreenDumpDir())
		})
	}
}
