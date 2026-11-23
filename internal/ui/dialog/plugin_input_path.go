// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// maxPathCompletions caps the number of entries offered for a path input.
const maxPathCompletions = 50

// completePath returns file system completion candidates for the given input
// text. Directory candidates are suffixed with a path separator so repeated
// completions drill down into them.
func completePath(text string) []string {
	if text == "" {
		return nil
	}

	var dir, prefix string
	if text == "~" {
		dir = "~" + string(os.PathSeparator)
	} else {
		dir, prefix = filepath.Split(text)
		if dir == "" {
			dir = "."
		}
	}

	base := expandTilde(dir)
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}

	completions := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		// Only suggest hidden entries when the user is explicitly looking for them.
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}

		completion := dir + name
		if isDir(base, entry) {
			completion += string(os.PathSeparator)
		}
		completions = append(completions, completion)
		if len(completions) >= maxPathCompletions {
			break
		}
	}

	return completions
}

func isDir(path string, entry fs.DirEntry) bool {
	if entry.IsDir() {
		return true
	}
	if entry.Type()&fs.ModeSymlink == 0 {
		return false
	}
	info, err := os.Stat(filepath.Join(path, entry.Name()))

	return err == nil && info.IsDir()
}

func expandTilde(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, strings.TrimPrefix(path, "~"))
}
