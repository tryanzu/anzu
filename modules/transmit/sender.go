package transmit

type Sender interface {
	Emit(channel, event string, params map[string]interface{})
}
