package mail

type Mail struct {
	Subject   string
	Template  int
	Recipient []MailRecipient
	Variables map[string]interface{}
	FromName  string
	FromEmail string
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
