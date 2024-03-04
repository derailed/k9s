package view

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fakeLookPath = func(file string) (string, error) {
	return fmt.Sprintf("/usr/bin/%s", file), nil
}

func TestFindEditor(t *testing.T) {
	type result struct {
		binary string
		args   []string
	}

	uu := map[string]struct {
		env map[string]string
		e   result
		err bool
	}{
		"no-editor": {
			env: map[string]string{},
			e:   result{binary: "", args: nil},
			err: true,
		},
		"vi-by-EDITOR": {
			env: map[string]string{"EDITOR": "vi"},
			e:   result{binary: "vi", args: nil},
		},
		"vim-with-args-by-EDITOR": {
			env: map[string]string{"EDITOR": "vim --wait --some=thing"},
			e:   result{binary: "vim", args: []string{"--wait", "--some=thing"}},
		},
		"code-with-args-by-KUBE_EDITOR": {
			env: map[string]string{"KUBE_EDITOR": "code -w"},
			e:   result{binary: "code", args: []string{"-w"}},
		},
		"code-with-args-by-K9S-EDITOR": {
			env: map[string]string{"K9S_EDITOR": "code --wait"},
			e:   result{binary: "code", args: []string{"--wait"}},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			opts := shellOpts{}

			for _, v := range editorEnvVars {
				if _, ok := u.env[v]; !ok {
					t.Setenv(v, "")
				}
			}

			for k, v := range u.env {
				t.Setenv(k, v)
			}

			got, err := findEditor(opts)
			assert.Equal(t, u.err, err != nil)
			assert.Contains(t, got.binary, u.e.binary)
			assert.Equal(t, u.e.args, got.args)
		})
	}
}
