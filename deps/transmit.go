package deps

import (
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
	zmq "github.com/pebbe/zmq4"
)

// Bootstraps transmit driver.
func IgniteTransmit(container Deps) (Deps, error) {
	transmitter := container.Config().UString("transmit.driver", "zmq")

	switch transmitter {
	case "zmq":
		address, err := container.Config().String("zmq.push")
		if err != nil {
			return container, err
		}

		//  Socket to send messages on
		sender, err := zmq.NewSocket(zmq.PUSH)
		if err != nil {
			return container, err
		}

		err = sender.Connect("tcp://" + address)
		if err != nil {
			return container, err
		}

		sender.Send("0", 0)

		container.TransmitProvider = transmit.ZMQ{sender, container.Log()}
	case "test":
		container.TransmitProvider = transmit.Test{container.Log()}
	}

	return container, nil
}
