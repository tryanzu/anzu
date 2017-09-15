package notifications

import (
	"github.com/fernandez14/spartangeek-blacker/deps"
)

// How much capacity each of the incoming notifications channels will have.
const BuffersLength = 10
const PoolSize = 2

// Define pool of channels.
var (
	Transmit chan Socket
	Database chan Notification
)

func init() {
	Transmit = make(chan Socket, BuffersLength)
	Database = make(chan Notification, BuffersLength)

	for n := 0; n < PoolSize; n++ {
		go databaseWorker(n, deps.Container)
		go transmitWorker(n)
	}
}
