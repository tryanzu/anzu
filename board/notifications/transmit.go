package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/deps"
)

func transmitWorker(n int) {
	for n := range Transmit {
		deps.Container.Transmit().Emit(n.Chan, n.Action, n.Params)
	}
}
