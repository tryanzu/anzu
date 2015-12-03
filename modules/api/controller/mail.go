package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/assets"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"io/ioutil"
	"encoding/json"
)

type MailAPI struct {
	Mongo   *mongo.Service `inject:""`
	Store   *store.Module  `inject:""`
	Assets  *assets.Module `inject:""`
	Errors  *exceptions.ExceptionsModule `inject:""`
}

// Get an inbound mail from mandrill
func (self MailAPI) Inbound(c *gin.Context) {

	address := c.Param("address")

	if address == "pc" {

		// Fallback to mandrill handler
		self.MandrillFallback(c)

		return
	}

	request := c.Request
	content, err := ioutil.ReadAll(request.Body)

	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Coulnt get payload."})
		return
	}

	var form MailInbound

	err = json.Unmarshal(content, &form)

	if err != nil {
		c.JSON(400, gin.H{"status": "error", "message": "Coulnt parse request, aborting.", "details": err})
		return
	}

	form.Id = bson.NewObjectId()

	go func(m MailInbound, self MailAPI) {

		// Recover from any panic even inside this goroutine
		defer self.Errors.Recover()
		
		assets := self.Assets
		database := self.Mongo.Database

		if len(m.Attachments) > 0 {

			for _, attachment := range m.Attachments {

				err := assets.UploadBase64(attachment.Content, attachment.Name, "mail", m.Id, attachment)

				if err != nil {
					panic(err)
				}
			}
		}

		err = database.C("inbound_mails").Insert(form)		

		if err != nil {
			panic(err)
		}

		order, err := self.Store.OrderFinder(m.FromFull.Email)

		if err == nil {

			var text string

			if len(m.StrippedTextReply) > 0 {

				text = m.StrippedTextReply

			} else {

				text = m.TextBody
			}

			// Push inbound answer
			order.PushInboundAnswer(text, m.Id)
		} else {

			// TODO - Forward email
		}

	}(form, self)
	
	c.JSON(200, gin.H{"status": "okay"})
}

func (self MailAPI) MandrillFallback(c *gin.Context) {

	var events []MailMandrillInbound

	mandrill_request := c.PostForm("mandrill_events")

	err := json.Unmarshal([]byte(mandrill_request), &events)

	if err != nil {

		c.JSON(400, gin.H{"status": "Could not process the request the way we expected."})
		return
	}

	// Due that we don't know how mandrill does stuff internally we'll respond back as fast as we can
	go func(list []MailMandrillInbound, self MailAPI) {

		database := self.Mongo.Database

		for _, mail := range list {

			// Assign an id to each mail so we can go further and relate them
			id := bson.NewObjectId()
			mail.Id = id

			// Persist each inbound email 
			err := database.C("mails").Insert(mail)

			if err != nil {
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
	Id        bson.ObjectId    `bson:"_id,omitempty" json:"id"`
	MessageID string           `json:"MessageID"`
	From      string           `json:"From"`
	FromName  string           `json:"FromName"`
	FromFull MailFullInbound   `json:"FromFull"`
	To       string            `json:"To"`
	ToFull   []MailFullInbound `json:"ToFull"`
	Cc       string            `json:"Cc"`
	CcFull   []MailFullInbound `json:"CcFull"`
	Bcc      string            `json:"Bcc"`
	BccFull  []MailFullInbound `json:"BccFull"`
	Subject  string            `json:"Subject"`
	Date     string            `json:"Date"`
	ReplyTo  string            `json:"ReplyTo"`
	MailboxHash  string        `json:"MailboxHash"`
	OriginalRecipient string   `json:"OriginalRecipient"`
	TextBody string   `json:"TextBody"`
	HtmlBody string   `json:"HtmlBody"`
	StrippedTextReply string `json:"StrippedTextReply"`

	Attachments  []MailAttachmentInbound `bson:"-" json:"Attachments"`
}

type MailFullInbound struct {
	Email       string `json:"Email"`
	Name        string `json:"Name"`
	MailboxHash string `json:"MailboxHash"`
}

type MailAttachmentInbound struct {
	Name    string `json:"Name"`
	Content string `json:"Content"`
	ContentType string `json:"ContentType"`
	ContentLength int `json:"ContentLength"`
	ContentID string `json:"ContentID"`
}

type MailMandrillInbound struct {
	Id        bson.ObjectId      `bson:"_id,omitempty" json:"id,omitempty"`
	Timestamp int                `json:"ts"`
	Event     string             `json:"event"`
	Message   MailMandrillInboundMessage `json:"msg"`
}

type MailMandrillInboundMessage struct {
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
