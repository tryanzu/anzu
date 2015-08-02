package exceptions

import (
	"errors"
	"fmt"
	"github.com/getsentry/raven-go"
)

type ExceptionsModule struct {
	ErrorService *raven.Client `inject:""`
}

func (di *ExceptionsModule) Recover() {

	var packet *raven.Packet

	switch rval := recover().(type) {
	case nil:
		return
	case error:
		packet = raven.NewPacket(rval.Error(), raven.NewException(rval, raven.NewStacktrace(2, 3, nil)))
	default:
		rvalStr := fmt.Sprint(rval)
		packet = raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.NewStacktrace(2, 3, nil)))
	}

	// Grab the error and send it to sentry
	di.ErrorService.Capture(packet, map[string]string{})
}
