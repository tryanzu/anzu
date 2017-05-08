package transmit

import (
	"encoding/json"
	"strings"
)

type Message struct {
	Room    string                 `json:"room"`
	Event   string                 `json:"event"`
	Message map[string]interface{} `json:"message"`
}

type Messages struct {
	List []map[string]interface{} `json:"list"`
}

func (m Message) RoomID() string {
	return strings.Replace(m.Event, " ", ":", -1)
}

func (m Message) Encode() string {
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}
