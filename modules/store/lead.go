package store

import (
	"errors"
	"strings"
	"time"

	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"gopkg.in/mgo.v2/bson"
)

var InvalidLeadAnswer = errors.New("Invalid lead answer type.")

type Lead struct {
	Id         bson.ObjectId   `bson:"_id,omitempty" json:"id,omitempty"`
	User       OrderUserModel  `bson:"user" json:"user"`
	Content    string          `bson:"content" json:"content"`
	Budget     int             `bson:"budget" json:"budget"`
	Currency   string          `bson:"currency" json:"currency"`
	State      string          `bson:"state" json:"state"`
	Usage      string          `bson:"usage" json:"usage"`
	Games      []string        `bson:"games" json:"games"`
	Extra      []string        `bson:"extras" json:"extra"`
	BuyDelay   int             `bson:"buydelay" json:"buydelay"`
	Unreaded   bool            `bson:"unreaded" json:"unreaded"`
	Messages   Messages        `bson:"messages,omitempty" json:"messages"`
	Tags       []TagModel      `bson:"tags,omitempty" json:"tags"`
	Activities []ActivityModel `bson:"activities,omitempty" json:"activities"`
	Pipeline   PipelineModel   `bson:"pipeline,omitempty" json:"pipeline"`
	Trusted    bool            `bson:"trusted_flag" json:"trusted_flag"`
	Favorite   bool            `bson:"favorite_flag" json:"favorite_flag"`
	Lead       bool            `bson:"-" json:"lead"`
	Created    time.Time       `bson:"created_at" json:"created_at"`
	Updated    time.Time       `bson:"updated_at" json:"updated_at"`

	// Runtime generated and not persisted in database
	RelatedUsers interface{}  `bson:"-" json:"related_users,omitempty"`
	Duplicates   []OrderModel `bson:"-" json:"duplicates,omitempty"`
	Invoice      *Invoice     `bson:"-" json:"invoice,omitempty"`

	deps Deps
}

// Reply logic over a lead.
func (lead *Lead) Reply(answer, kind string) error {
	db := lead.deps.Mgo()
	mailer := lead.deps.Mailer()

	if kind != "text" && kind != "note" {
		return InvalidLeadAnswer
	}

	var id, subject string
	if kind == "text" {
		answer = strings.Replace(answer, "\n", "<br>", -1)
		subject = "PC Spartana"

		if len(lead.Messages) > 0 {
			subject = "RE: " + subject
		}

		compose := mail.Mail{
			Subject:  subject,
			Template: 250241,
			Recipient: []mail.MailRecipient{
				{
					Name:  lead.User.Name,
					Email: lead.User.Email,
				},
			},
			FromEmail: "pc@spartangeek.com",
			FromName:  "Drak Spartan",
			Variables: map[string]interface{}{
				"content": answer,
				"subject": subject,
			},
		}

		id = mailer.Send(compose)
	}

	message := MessageModel{
		Content:   answer,
		Type:      kind,
		MessageID: id,
		Created:   time.Now(),
		Updated:   time.Now(),
	}

	err := db.C("orders").Update(
		bson.M{"_id": lead.Id},
		bson.M{
			"$push": bson.M{"messages": message},
			"$set":  bson.M{"updated_at": time.Now()},
		})

	if err != nil {
		return err
	}

	return nil
}
