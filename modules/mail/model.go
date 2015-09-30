package mail

type Mail struct {
	Subject   string
	Template  string
	Recipient []MailRecipient
	Variables map[string]string
}

type MailRecipient struct {
	Name  string
	Email string
}

type ModuleConfig struct {
	From       string
	FromName   string
	Recipients []string
}
