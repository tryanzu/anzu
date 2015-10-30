package controller

import (
	"encoding/json"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"log"
)

type MailAPI struct {
	Mongo *mongo.Service `inject:""`
	Store *store.Module  `inject:""`
}

// Get an inbound mail from mandrill
func (self MailAPI) Inbound(c *gin.Context) {

	var events []MailInbound

	// Even though is weird the json comes into a form
	mandrill_request := c.PostForm("mandrill_events")

	err := json.Unmarshal([]byte(mandrill_request), &events)

	if err != nil {

		c.JSON(400, gin.H{"status": "Could not process the request the way we expected."})
		return
	}

	// Due that we don't know how mandrill does stuff internally we'll respond back as fast as we can
	go func(list []MailInbound, self MailAPI) {

		database := self.Mongo.Database

		for _, mail := range list {

			// Assign an id to each mail so we can go further and relate them
			id := bson.NewObjectId()
			mail.Id = id

			// Persist each inbound email 
			err := database.C("mails").Insert(mail)

			if err != nil {

				log.Printf("[error] %v\n", err.Error())

				continue
			}

			order, err := self.Store.OrderFinder(mail.Message.FromEmail)

			if err == nil {

				order.PushInboundAnswer(mail.Message.Text, id)
			}
		}

	}(events, self)

	c.JSON(200, gin.H{"status": "okay"})
}

type MailInbound struct {
	Id        bson.ObjectId      `bson:"_id,omitempty" json:"id,omitempty"`
	Timestamp int                `json:"ts"`
	Event     string             `json:"event"`
	Message   MailInboundMessage `json:"msg"`
}

type MailInboundMessage struct {
	Headers   map[string]interface{} `json:"headers"`
	Text      string                 `json:"text"`
	FromEmail string                 `json:"from_email"`
	FromName  string                 `json:"from_name"`
	To        interface{}            `json:"to"`
	Spam      interface{}            `json:"spam_report"`
	Email     string                 `json:"email"`
	Subject   string                 `json:"subject"`
	Raw       interface{}            `json:"raw_msg"`
	Dkim      interface{}            `json:"dkim"`
	Spf       interface{}            `json:"spf"`
	Sender    interface{}            `json:"sender"`
	Tags      interface{}            `json:"tags"`
	Html      string                 `json:"html"`
	Flowed    interface{}            `json:"text_flowed"`
	Template  interface{}            `json:"template"`
}
