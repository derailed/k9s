package render

import (
	"io/ioutil"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestAugmentRow(t *testing.T) {
	uu := map[string]struct {
		file string
		e    Fields
	}{
		"cool": {
			"assets/b1.txt",
			Fields{"pass", "3.3544", "29.8116", "100", "0"},
		},
		"2XX": {
			"assets/b4.txt",
			Fields{"pass", "3.3544", "29.8116", "160", "0"},
		},
		"4XX/5XX": {
			"assets/b2.txt",
			Fields{"pass", "3.3544", "29.8116", "100", "12"},
		},
		"toast": {
			"assets/b3.txt",
			Fields{"fail", "2.3688", "35.4606", "0", "0"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			data, err := ioutil.ReadFile(u.file)

			assert.Nil(t, err)
			fields := make(Fields, 8)
			b := Benchmark{}
			b.augmentRow(fields, string(data))
			assert.Equal(t, u.e, fields[2:7])
		})
	}
}
