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

func (t TestMailer) SendRaw(raw Raw) string {
	t.Logger.Debugf("Mail send: %+v", raw)
	return "test-id"
}
