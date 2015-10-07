package store

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"gopkg.in/mgo.v2/bson"
	"time"
	"strings"
)

type One struct {
	di   *Module
	data *OrderModel
}

func (self *One) Data() *OrderModel {

	return self.data
}

func (self *One) PushAnswer(text, kind string) {

	if kind != "text" && kind != "note" {

		return
	}

	database := self.di.Mongo.Database

	message := MessageModel{
		Content: text,
		Type:    kind,
		Created: time.Now(),
		Updated: time.Now(),
	}

	err := database.C("orders").Update(bson.M{"_id": self.data.Id}, bson.M{"$push": bson.M{"messages": message}, "$set": bson.M{"updated_at": time.Now()}})

	if err != nil {
		panic(err)
	}

	if kind == "text" {

		// Send an email async
		go func() {

			mailing := self.di.Mail
			text = strings.Replace(text, "\n", "<br>", -1)
			
			compose := mail.Mail{
				Subject:  "PC Spartana",
				Template: "simple",
				Recipient: []mail.MailRecipient{
					{
						Name:  self.data.User.Name,
						Email: self.data.User.Email,
					},
				},
				FromEmail: "pc@pedidos.spartangeek.com",
				FromName: "Drak Spartan",
				Variables: map[string]string{
					"content": text,
				},
			}

			mailing.Send(compose)
		}()
	}
}

func (self *One) PushInboundAnswer(text string, mail bson.ObjectId) {

	database := self.di.Mongo.Database

	message := MessageModel{
		Content: text,
		Type:    "inbound",
		RelatedId: mail,
		Created: time.Now(),
		Updated: time.Now(),
	}

	err := database.C("orders").Update(bson.M{"_id": self.data.Id}, bson.M{"$push": bson.M{"messages": message}, "$set": bson.M{"updated_at": time.Now()}})

	if err != nil {
		panic(err)
	}
}