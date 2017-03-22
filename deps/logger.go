package deps

import (
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("blacker")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func IgniteLogger(container Deps) (Deps, error) {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(formatter)
	container.Logger = log
	return container, nil
}
