package view

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/derailed/tview"

	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog/log"
)

var (
	keyValRX = regexp.MustCompile(`\A(\s*)([\w|\-|\.|\/|\s]+):\s(.+)\z`)
	keyRX    = regexp.MustCompile(`\A(\s*)([\w|\-|\.|\/|\s]+):\s*\z`)
)

const (
	yamlFullFmt  = "%s[key::b]%s[colon::-]: [val::]%s"
	yamlKeyFmt   = "%s[key::b]%s[colon::-]:"
	yamlValueFmt = "[val::]%s"
)

func colorizeYAML(style config.Yaml, raw string) string {
	lines := strings.Split(tview.Escape(raw), "\n")

	fullFmt := strings.Replace(yamlFullFmt, "[key", "["+style.KeyColor, 1)
	fullFmt = strings.Replace(fullFmt, "[colon", "["+style.ColonColor, 1)
	fullFmt = strings.Replace(fullFmt, "[val", "["+style.ValueColor, 1)

	keyFmt := strings.Replace(yamlKeyFmt, "[key", "["+style.KeyColor, 1)
	keyFmt = strings.Replace(keyFmt, "[colon", "["+style.ColonColor, 1)

	valFmt := strings.Replace(yamlValueFmt, "[val", "["+style.ValueColor, 1)

	buff := make([]string, 0, len(lines))
	for _, l := range lines {
		res := keyValRX.FindStringSubmatch(l)
		if len(res) == 4 {
			buff = append(buff, fmt.Sprintf(fullFmt, res[1], res[2], res[3]))
			continue
		}

		res = keyRX.FindStringSubmatch(l)
		if len(res) == 3 {
			buff = append(buff, fmt.Sprintf(keyFmt, res[1], res[2]))
			continue
		}

		buff = append(buff, fmt.Sprintf(valFmt, l))
	}

	return strings.Join(buff, "\n")
}

func saveYAML(cluster, name, data string) (string, error) {
	dir := filepath.Join(config.K9sDumpDir, cluster)
	if err := ensureDir(dir); err != nil {
		return "", err
	}

	now := time.Now().UnixNano()
	fName := fmt.Sprintf("%s-%d.yml", strings.Replace(name, "/", "-", -1), now)

	path := filepath.Join(dir, fName)
	mod := os.O_CREATE | os.O_WRONLY
	file, err := os.OpenFile(path, mod, 0600)
	if err != nil {
		log.Error().Err(err).Msgf("YAML create %s", path)
		return "", nil
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error().Err(err).Msg("Closing yaml file")
		}
	}()
	if _, err := file.Write([]byte(data)); err != nil {
		return "", err
	}

	return path, nil
}
