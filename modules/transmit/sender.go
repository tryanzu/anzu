package transmit

import (
	zmq "github.com/pebbe/zmq4"

	"encoding/json"
)

type Sender struct {
	socket *zmq.Socket
}

func (this *Sender) Emit(channel, event string, params map[string]interface{}) {

	room := channel
	roomEvent := room + " " + event

	message := Message{
		Room: channel,
		Event: roomEvent,
		Message: params,
	}

	msg, err := json.Marshal(message)

	if err != nil {
		panic(err)
	}

	this.socket.Send(string(msg), zmq.DONTWAIT)
}