// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

// MockSADir overrides the service account mount dir for tests.
func MockSADir(dir string) (undo func()) {
	old := saDir
	saDir = dir

	return func() { saDir = old }
}
