// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package slogs

import "log/slog"

// CLog returns a child logger.
func CLog(subsys string) *slog.Logger {
	return slog.With(Subsys, subsys)
}
