//go:build linux

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func init() {
	// Ensure that when the parent process (e.g. bash / sshd) terminates,
	// the Linux kernel delivers SIGHUP to k9s so it doesn't become an orphaned process.
	_ = unix.Prctl(unix.PR_SET_PDEATHSIG, uintptr(syscall.SIGHUP), 0, 0, 0)
	if syscall.Getppid() == 1 {
		// If the parent already died before prctl was called, self-terminate immediately.
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	}
}
