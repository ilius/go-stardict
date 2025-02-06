package stardict

import "log/slog"

var ErrorHandler = func(err error) {
	slog.Error("error", "err", err)
}
