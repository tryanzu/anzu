package builds

import (
	"github.com/fernandez14/spartangeek-blacker/modules/builds"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"gopkg.in/jmcvetta/neoism.v1"
)

type API struct {
	Components *components.Module `inject:""`
	Neoism     *neoism.Database   `inject:""`
	Builds     *builds.Module     `inject:""`
}
