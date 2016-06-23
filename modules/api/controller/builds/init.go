package builds

import (
	"github.com/fernandez14/spartangeek-blacker/modules/builds"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
)

type API struct {
	Components *components.Module `inject:""`
	Builds     *builds.Module     `inject:""`
}
