// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

//go:build windows

package termtitle

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// maxTitle is the longest console title Windows accepts, in UTF-16 code units.
const maxTitle = 32 * 1024

var (
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")
	procGetConsoleTW = kernel32.NewProc("GetConsoleTitleW")
	procSetConsoleTW = kernel32.NewProc("SetConsoleTitleW")
)

// consoleBackend uses the console API, which works in conhost and Windows Terminal.
type consoleBackend struct {
	original string
}

func newBackend() backend {
	return &consoleBackend{}
}

func (b *consoleBackend) save() error {
	if err := procGetConsoleTW.Find(); err != nil {
		return err
	}
	buf := make([]uint16, maxTitle)
	n, _, err := procGetConsoleTW.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 && !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	b.original = windows.UTF16ToString(buf[:n])

	return nil
}

func (*consoleBackend) set(title string) error {
	return setConsoleTitle(title)
}

func (b *consoleBackend) restore() error {
	return setConsoleTitle(b.original)
}

func setConsoleTitle(title string) error {
	p, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return err
	}
	if r, _, err := procSetConsoleTW.Call(uintptr(unsafe.Pointer(p))); r == 0 {
		return err
	}

	return nil
}
