// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"os"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func Test_activeConfig(t *testing.T) {
	os.Setenv(config.K9sEnvConfigDir, "/tmp/test-config")
	assert.NoError(t, config.InitLocs())

	cl, ct := "cl-1", "ct-1-1"
	uu := map[string]struct {
		cl, ct string
		cfg    *Configurator
		ok     bool
	}{
		"empty": {
			cfg: &Configurator{},
		},

		"plain": {
			cfg: &Configurator{Config: config.NewConfig(
				mock.NewMockKubeSettings(&genericclioptions.ConfigFlags{
					ClusterName: &cl,
					Context:     &ct,
				}))},
			cl: cl,
			ct: ct,
			ok: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cfg := u.cfg
			if cfg.Config != nil {
				_, err := cfg.Config.K9s.ActivateContext(ct)
				assert.NoError(t, err)
			}
			cl, ct, ok := cfg.activeConfig()
			assert.Equal(t, u.ok, ok)
			if ok {
				assert.Equal(t, u.cl, cl)
				assert.Equal(t, u.ct, ct)
			}
		})
	}
}
