package view

import (
	"io/ioutil"
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestAugmentRow(t *testing.T) {
	uu := map[string]struct {
		file string
		e    render.Fields
	}{
		"cool": {
			"test_assets/b1.txt",
			render.Fields{"pass", "3.3544", "29.8116", "100", "0"},
		},
		"2XX": {
			"test_assets/b4.txt",
			render.Fields{"pass", "3.3544", "29.8116", "160", "0"},
		},
		"4XX/5XX": {
			"test_assets/b2.txt",
			render.Fields{"pass", "3.3544", "29.8116", "100", "12"},
		},
		"toast": {
			"test_assets/b3.txt",
			render.Fields{"fail", "2.3688", "35.4606", "0", "0"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			data, err := ioutil.ReadFile(u.file)

			assert.Nil(t, err)
			fields := make(render.Fields, 8)
			augmentRow(fields, string(data))
			assert.Equal(t, u.e, fields[2:7])
		})
	}
}
