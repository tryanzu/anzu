package mail

import (
	"gopkg.in/mgo.v2/bson"

	"time"
)

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
