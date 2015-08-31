package core

import (
	"github.com/facebookgo/inject"
)

type CoreModule struct {
	Dependencies []*inject.Object
}

func (core *CoreModule) Inject(service *inject.Object) {

	core.Dependencies = append(core.Dependencies, service)
}
