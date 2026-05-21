// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package termdetect

import (
	"fmt"
	"os"

	"github.com/derailed/tcell/v2"
	"github.com/lucasb-eyer/go-colorful"
	"gopkg.in/yaml.v3"
)

func SkinIsLight(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	var s struct {
		K9s struct {
			Body struct {
				BgColor string `yaml:"bgColor"`
			} `yaml:"body"`
		} `yaml:"k9s"`
	}
	if err := yaml.Unmarshal(data, &s); err != nil {
		return false, err
	}
	if s.K9s.Body.BgColor == "" || s.K9s.Body.BgColor == "-" || s.K9s.Body.BgColor == "default" {
		return false, nil
	}
	tc := tcell.GetColor(s.K9s.Body.BgColor).TrueColor()
	hex := tc.Hex()
	if hex < 0 {
		return false, nil
	}
	col, err := colorful.Hex(fmt.Sprintf("#%06x", hex))
	if err != nil {
		return false, nil
	}
	return luminance(col) > 0.5, nil
}
