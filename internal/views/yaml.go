package views

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	keyValRX = regexp.MustCompile(`\A(\s*)([\w|\-|\.|\s]+):\s(.+)\z`)
	keyRX    = regexp.MustCompile(`\A(\s*)([\w|\-|\.|\s]+):\s*\z`)
)

func colorizeYAML(raw string) string {
	lines := strings.Split(raw, "\n")

	buff := make([]string, 0, len(lines))
	for _, l := range lines {
		res := keyValRX.FindStringSubmatch(l)
		if len(res) == 4 {
			buff = append(buff, fmt.Sprintf("%s[steelblue::b]%s[white::-]: [papayawhip::]%s", res[1], res[2], res[3]))
			continue
		}

		res = keyRX.FindStringSubmatch(l)
		if len(res) == 3 {
			buff = append(buff, fmt.Sprintf("%s[steelblue::b]%s[white::-]:", res[1], res[2]))
			continue
		}

		buff = append(buff, fmt.Sprintf("[papayawhip::]%s", l))
	}

	return strings.Join(buff, "\n")
}
