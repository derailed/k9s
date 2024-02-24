// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"os"
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestAugmentRow(t *testing.T) {
	uu := map[string]struct {
		file string
		e    model1.Fields
	}{
		"cool": {
			"testdata/b1.txt",
			model1.Fields{"pass", "3.3544", "29.8116", "100", "0"},
		},
		"2XX": {
			"testdata/b4.txt",
			model1.Fields{"pass", "3.3544", "29.8116", "160", "0"},
		},
		"4XX/5XX": {
			"testdata/b2.txt",
			model1.Fields{"pass", "3.3544", "29.8116", "100", "12"},
		},
		"toast": {
			"testdata/b3.txt",
			model1.Fields{"fail", "2.3688", "35.4606", "0", "0"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			data, err := os.ReadFile(u.file)

			assert.Nil(t, err)
			fields := make(model1.Fields, 8)
			b := Benchmark{}
			b.augmentRow(fields, string(data))
			assert.Equal(t, u.e, fields[2:7])
		})
	}
}
