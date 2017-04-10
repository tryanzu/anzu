package transmit

import (
	"github.com/op/go-logging"
)

type Test struct {
	Logger *logging.Logger
}

func (pool Test) Emit(channel, event string, params map[string]interface{}) {
	pool.Logger.Debugf("Sending %s to %s with %+v", event, channel, params)
}
