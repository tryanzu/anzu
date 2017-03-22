package mail

type Mailer interface {
	Send(Mail) string
}
