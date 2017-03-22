package mail

type Mailer interface {
	Send(Mail) string
	SendRaw(Raw) string
}
