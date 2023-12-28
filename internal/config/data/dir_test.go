// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data_test

import (
	"os"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestDirLoad(t *testing.T) {
	uu := map[string]struct {
		dir   string
		flags *genericclioptions.ConfigFlags
		err   error
		cfg   *data.Config
	}{
		"happy-cl-1-ct-1": {
			dir:   "testdata/data/k9s",
			flags: makeFlags("cl-1", "ct-1-1"),
			cfg:   mustLoadConfig("testdata/configs/ct-1-1.yaml"),
		},

		"happy-cl-1-ct2": {
			dir:   "testdata/data/k9s",
			flags: makeFlags("cl-1", "ct-1-2"),
			cfg:   mustLoadConfig("testdata/configs/ct-1-2.yaml"),
		},

		"happy-cl-2": {
			dir:   "testdata/data/k9s",
			flags: makeFlags("cl-2", "ct-2-1"),
			cfg:   mustLoadConfig("testdata/configs/ct-2-1.yaml"),
		},

		"toast": {
			dir:   "/tmp/data/k9s",
			flags: makeFlags("cl-test", "ct-test-1"),
			cfg:   mustLoadConfig("testdata/configs/def_ct.yaml"),
		},

		"non-sanitized-path": {
			dir:   "/tmp/data/k9s",
			flags: makeFlags("arn:aws:eks:eu-central-1:xxx:cluster/fred-blee", "fred-blee"),
			cfg:   mustLoadConfig("testdata/configs/aws_ct.yaml"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.NotNil(t, u.cfg, "test config must not be nil")
			if u.cfg == nil {
				return
			}

			ks := mock.NewMockKubeSettings(u.flags)
			if strings.Index(u.dir, "/tmp") == 0 {
				assert.NoError(t, mock.EnsureDir(u.dir))
			}

			d := data.NewDir(u.dir)
			ct, err := ks.CurrentContext()
			assert.NoError(t, err)
			if err != nil {
				return
			}

			cfg, err := d.Load(*u.flags.Context, ct)
			assert.Equal(t, u.err, err)
			if u.err == nil {
				assert.Equal(t, u.cfg, cfg)
			}
		})
	}
}

// Helpers...

func makeFlags(cl, ct string) *genericclioptions.ConfigFlags {
	return &genericclioptions.ConfigFlags{
		ClusterName: &cl,
		Context:     &ct,
	}
}

func mustLoadConfig(cfg string) *data.Config {
	bb, err := os.ReadFile(cfg)
	if err != nil {
		return nil
	}
	var ct data.Config
	if err = yaml.Unmarshal(bb, &ct); err != nil {
		return nil
	}

	return &ct
}
