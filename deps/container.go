package deps

import (
	slog "log"
	"reflect"
)

// Contains bootstraped dependencies.
var (
	Container Deps
)

// An ignitor takes a Container and injects bootstraped dependencies.
type Ignitor func(Deps) (Deps, error)

// Runs ignitors to fulfill deps container.
func Bootstrap() {
	ignitors := []Ignitor{
		IgniteLogger,
		IgniteConfig,
		IgniteMongoDB,
		IgniteMailer,
	}

	var err error
	Container = Deps{}
	for _, fn := range ignitors {
		Container, err = fn(Container)
		if err != nil {
			slog.Panic(err)
		}
	}

	v := reflect.ValueOf(Container)
	fields := reflect.Indirect(v)

	for i := 0; i < v.NumField(); i++ {
		slog.Printf("Bootstraped %s", fields.Type().Field(i).Name)
	}
}
