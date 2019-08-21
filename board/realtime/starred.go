package realtime

import (
	"encoding/json"
	"time"
)

var featuredM chan M

func starredMessagesWorker() {
	for m := range featuredM {
		var ev SocketEvent
		err := json.Unmarshal([]byte(m.Content), &ev)
		if err != nil {
			continue
		}
		ev.Params["at"] = time.Now()
		m.Content = ev.encode()
		ToChan <- m
		time.Sleep(15 * time.Second)
	}
}
