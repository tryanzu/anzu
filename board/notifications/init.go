package notifications

// How much capacity each of the incoming notifications channels will have.
const BuffersLength = 10
const PoolSize = 4

// Define pool of channels.
var (
	Transmit chan Socket
	Database chan Notification
)

func init() {
	Transmit = make(chan Socket, BuffersLength)
	Database = make(chan Notification, BuffersLength)

	for n := 0; n < PoolSize; n++ {
		go databaseWorker(n)
		go transmitWorker(n)
	}
}
