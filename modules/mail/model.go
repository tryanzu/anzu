package mail

import (
	"gopkg.in/mgo.v2/bson"

	"io"
	core "net/mail"
	"time"
)

type Mailable interface {
	From() *core.Address
	To() []*core.Address
	SubjectText() string
}

type MailBase struct {
	Subject   string
	Recipient []MailRecipient
	Variables map[string]interface{}
	FromName  string
	FromEmail string
}

func (mail MailBase) SubjectText() string {
	return mail.Subject
}

func (mail MailBase) From() *core.Address {
	return &core.Address{Name: mail.FromName, Address: mail.FromEmail}
}

func (mail MailBase) To() []*core.Address {
	var recipients []*core.Address
	for _, r := range mail.Recipient {
		recipients = append(recipients, &core.Address{
			Name:    r.Name,
			Address: r.Email,
		})
	}

	return recipients
}

type Mail struct {
	MailBase
	Template int
}

type Raw struct {
	MailBase
	Content io.Reader
}

type MailRecipient struct {
	Name  string
	Email string
}

type ModuleConfig struct {
	From           string
	FromName       string
	Recipients     []string
	IgnoredDomains []string
}

type InboundMail struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id"`
	MessageId string        `bson:"messageid" json:"message_id"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
}
