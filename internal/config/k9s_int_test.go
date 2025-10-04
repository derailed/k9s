// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_k9sOverrides(t *testing.T) {
	var (
		trueVal = true
		cmd     = "po"
		dir     = "/tmp/blee"
	)

	uu := map[string]struct {
		k                  *K9s
		rate               float32
		ro, hl, cl, sl, ll bool
	}{
		"plain": {
			k: &K9s{
				LiveViewAutoRefresh: false,
				ScreenDumpDir:       "",
				RefreshRate:         10.0,
				MaxConnRetry:        0,
				ReadOnly:            false,
				NoExitOnCtrlC:       false,
				UI:                  UI{},
				SkipLatestRevCheck:  false,
				DisablePodCounting:  false,
			},
			rate: 10.0,
		},
		"sub-second": {
			k: &K9s{
				LiveViewAutoRefresh: false,
				ScreenDumpDir:       "",
				RefreshRate:         0.5,
				MaxConnRetry:        0,
				ReadOnly:            false,
				NoExitOnCtrlC:       false,
				UI:                  UI{},
				SkipLatestRevCheck:  false,
				DisablePodCounting:  false,
			},
			rate: 2.0, // minimum enforced
		},
		"set": {
			k: &K9s{
				LiveViewAutoRefresh: false,
				ScreenDumpDir:       "",
				RefreshRate:         10.0,
				MaxConnRetry:        0,
				ReadOnly:            true,
				NoExitOnCtrlC:       false,
				UI: UI{
					Headless:   true,
					Logoless:   true,
					Crumbsless: true,
					Splashless: true,
				},
				SkipLatestRevCheck: false,
				DisablePodCounting: false,
			},
			rate: 10.0,
			ro:   true,
			hl:   true,
			ll:   true,
			cl:   true,
			sl:   true,
		},
		"overrides": {
			k: &K9s{
				LiveViewAutoRefresh: false,
				ScreenDumpDir:       "",
				RefreshRate:         10.0,
				MaxConnRetry:        0,
				ReadOnly:            false,
				NoExitOnCtrlC:       false,
				UI: UI{
					Headless:         false,
					Logoless:         false,
					Crumbsless:       false,
					manualHeadless:   &trueVal,
					manualLogoless:   &trueVal,
					manualCrumbsless: &trueVal,
					manualSplashless: &trueVal,
				},
				SkipLatestRevCheck:  false,
				DisablePodCounting:  false,
				manualRefreshRate:   100.0,
				manualReadOnly:      &trueVal,
				manualCommand:       &cmd,
				manualScreenDumpDir: &dir,
			},
			rate: 100.0,
			ro:   true,
			hl:   true,
			ll:   true,
			cl:   true,
			sl:   true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.InDelta(t, u.rate, u.k.GetRefreshRate(), 0.001)
			assert.Equal(t, u.ro, u.k.IsReadOnly())
			assert.Equal(t, u.cl, u.k.IsCrumbsless())
			assert.Equal(t, u.sl, u.k.IsSplashless())
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
			require.NoError(t, cfg.Load("testdata/configs/k9s.yaml", true))

			cfg.K9s.manualScreenDumpDir = &u.dir
			assert.Equal(t, u.e, cfg.K9s.AppScreenDumpDir())
		})
	}
}
