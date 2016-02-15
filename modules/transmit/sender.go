package transmit

import (
	zmq "github.com/pebbe/zmq4"

	"encoding/json"
	"time"
)

type Sender struct {
	port string
}

func (this *Sender) Emit(channel, event string, params map[string]interface{}) {

	//  Socket to send messages on
	sender, err := zmq.NewSocket(zmq.PUSH)
	defer sender.Close()

	if err != nil {
		panic(err)
	}

	err = sender.Connect("tcp://127.0.0.1:" + this.port)

	if err != nil {
		panic(err)
	}

	sender.Send("0", 0)

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

	_, err = sender.Send(string(msg), 0)

	if err != nil {
		panic(err)
	}

	// Give 0mq time to deliver
	time.Sleep(time.Second)
}