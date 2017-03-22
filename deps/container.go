package deps

import (
	slog "log"
)

// Contains bootstraped dependencies.
var Container Deps

// An ignitor takes a Container and injects bootstraped dependencies.
type Ignitor func(Deps) (Deps, error)

// Runs ignitors to fulfill deps container.
func Bootstrap() {
	ignitors := []Ignitor{
		IgniteLogger,
		IgniteConfig,
		IgniteMongoDB,
	}

	Container = Deps{}
	for _, fn := range ignitors {
		Container, err := fn(Container)
		if err != nil {
			slog.Panic(err)
		}
	}
}
