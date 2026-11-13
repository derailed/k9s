// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package termtitle

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeBackend struct {
	calls   []string
	saveErr error
}

func (f *fakeBackend) save() error {
	f.calls = append(f.calls, "save")
	return f.saveErr
}

func (f *fakeBackend) set(title string) error {
	f.calls = append(f.calls, "set:"+title)
	return nil
}

func (f *fakeBackend) restore() error {
	f.calls = append(f.calls, "restore")
	return nil
}

func TestTitleLifecycle(t *testing.T) {
	b := new(fakeBackend)
	tt := &Title{backend: b}

	tt.Set("k9s | a")
	tt.Set("k9s | a")
	tt.Set("k9s | b")
	tt.Restore()
	tt.Restore()

	assert.Equal(t, []string{"save", "set:k9s | a", "set:k9s | b", "restore"}, b.calls)
}

func TestTitleRestoreWithoutSet(t *testing.T) {
	b := new(fakeBackend)
	tt := &Title{backend: b}

	tt.Restore()

	assert.Empty(t, b.calls)
}

func TestTitleSaveFails(t *testing.T) {
	b := &fakeBackend{saveErr: errors.New("no tty")}
	tt := &Title{backend: b}

	tt.Set("k9s")
	tt.Restore()

	assert.Equal(t, []string{"save"}, b.calls)
}

func TestSanitize(t *testing.T) {
	assert.Equal(t, "k9s | evil]0;x", Sanitize("k9s | evil\x1b]0;x\x07"))
	assert.Equal(t, "k9s | ctx", Sanitize("k9s | ctx\n"))
}
