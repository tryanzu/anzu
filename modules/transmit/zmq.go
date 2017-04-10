package transmit

import (
	"github.com/op/go-logging"
	zmq "github.com/pebbe/zmq4"

	"encoding/json"
	"time"
)

type ZMQ struct {
	Address string
	Logger  *logging.Logger
}

func (pool ZMQ) Emit(channel, event string, params map[string]interface{}) {

	//  Socket to send messages on
	sender, err := zmq.NewSocket(zmq.PUSH)
	defer sender.Close()

	if err != nil {
		panic(err)
	}

	err = sender.Connect("tcp://" + pool.Address)
	if err != nil {
		panic(err)
	}

	sender.Send("0", 0)

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

	_, err = sender.Send(string(msg), 0)

	if err != nil {
		panic(err)
	}

	// Give 0mq time to deliver
	time.Sleep(time.Second)
}
