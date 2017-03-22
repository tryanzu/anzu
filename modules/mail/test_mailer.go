package mail

import (
	"github.com/op/go-logging"
)

type TestMailer struct {
	Logger *logging.Logger
}

func (t TestMailer) Send(mail Mail) string {
	t.Logger.Debugf("Mail sent: %+v", mail)
	return "test-id"
}
