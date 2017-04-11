package transmit

import (
	"strings"
)

type Message struct {
	Room    string                 `json:"room"`
	Event   string                 `json:"event"`
	Message map[string]interface{} `json:"message"`
}

func (m Message) RoomID() string {
	return strings.Replace(m.Event, " ", ":", -1)
}
