package cmd

import "log/slog"

// slogWarn is a thin wrapper to avoid repetitive attribute construction at call sites.
func slogWarn(msg string, args ...any) {
	slog.Warn(msg, args...)
}
