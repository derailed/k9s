package views

import (
	"io/ioutil"
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestAugmentRow(t *testing.T) {
	uu := map[string]struct {
		file string
		e    resource.Row
	}{
		"cool": {
			"test_assets/b1.txt",
			resource.Row{"pass", "3.3544", "29.8116", "100", "0"},
		},
		"2XX": {
			"test_assets/b4.txt",
			resource.Row{"pass", "3.3544", "29.8116", "160", "0"},
		},
		"4XX/5XX": {
			"test_assets/b2.txt",
			resource.Row{"pass", "3.3544", "29.8116", "100", "12"},
		},
		"toast": {
			"test_assets/b3.txt",
			resource.Row{"fail", "2.3688", "35.4606", "0", "0"},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			data, err := ioutil.ReadFile(u.file)

			assert.Nil(t, err)
			fields := make(resource.Row, 8)
			augmentRow(fields, string(data))
			assert.Equal(t, u.e, fields[2:7])
		})
	}
}
