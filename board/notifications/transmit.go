package notifications

import (
	"encoding/json"

	"github.com/tryanzu/core/board/realtime"
)

func transmitWorker(n int) {
	for n := range Transmit {
		m, err := json.Marshal(n)
		if err != nil {
			panic(err)
		}

		realtime.Broadcast <- string(m)
	}
}
