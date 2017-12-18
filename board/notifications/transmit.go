package notifications

import (
	"encoding/json"

	"github.com/fernandez14/spartangeek-blacker/board/realtime"
	"github.com/fernandez14/spartangeek-blacker/deps"
)

func transmitWorker(n int) {
	for n := range Transmit {
		deps.Container.Transmit().Emit(n.Chan, n.Action, n.Params)

		m, err := json.Marshal(n)
		if err != nil {
			panic(err)
		}

		realtime.Broadcast <- string(m)
	}
}
