// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetailsSaveSanitize(t *testing.T) {
	// colorizeYAML runs the content through tview.Escape, turning each `[X]` into
	// `[X[]`. saveCmd must reverse that with sanitizeEsc, otherwise the file
	// written to disk keeps the escaped brackets (e.g. `[[:space:[]]`) and is no
	// longer valid for the sed/grep consumers the YAML came from. A line with
	// several POSIX bracket classes is the reproducer from #4042.
	s := `apiVersion: v1
data:
  script: sed -E 's/^([[:space:]]*count:[[:space:]]*)[0-9]+/x/'
`

	app := NewApp(mock.NewMockConfig(t))
	app.Config.K9s.ScreenDumpDir = t.TempDir()

	d := NewDetails(app, "fred", "subject", "yaml", false)
	require.NoError(t, d.Init(context.Background()))
	d.text.SetText(colorizeYAML(config.Yaml{}, s))

	require.Nil(t, d.saveCmd(nil))

	dir := app.Config.K9s.ContextScreenDumpDir()
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	require.NoError(t, err)
	require.Len(t, files, 1)

	got, err := os.ReadFile(files[0])
	require.NoError(t, err)
	assert.Equal(t, s, string(got))
}
