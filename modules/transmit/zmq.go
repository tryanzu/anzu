package transmit

import (
	"github.com/op/go-logging"
	zmq "github.com/pebbe/zmq4"

	"encoding/json"
)

type ZMQ struct {
	Socket *zmq.Socket
	Logger *logging.Logger
}

func (pool ZMQ) Emit(channel, event string, params map[string]interface{}) {

	room := channel
	roomEvent := room + " " + event
	message := Message{
		Room:    channel,
		Event:   roomEvent,
		Message: params,
	}

	msg, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	_, err = pool.Socket.SendMessageDontwait(string(msg), 0)
	if err != nil {
		panic(err)
	}
}
