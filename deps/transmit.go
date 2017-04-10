package deps

import (
	"github.com/fernandez14/spartangeek-blacker/modules/transmit"
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

		container.TransmitProvider = transmit.ZMQ{address, container.Log()}
	case "test":
		container.TransmitProvider = transmit.Test{container.Log()}
	}

	return container, nil
}
