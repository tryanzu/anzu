package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/assets"
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"encoding/json"
	"io/ioutil"
	"time"
)

type MailAPI struct {
	Mongo  *mongo.Service               `inject:""`
	Store  *store.Module                `inject:""`
	Assets *assets.Module               `inject:""`
	Errors *exceptions.ExceptionsModule `inject:""`
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

		// Save the time
		form.Created = time.Now()

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

func (m MailAPI) BounceWebhook(c *gin.Context) {

	var payload PostmarkBouncedPayload

	if c.Bind(&payload) == nil {

		db := m.Mongo.Database
		err := db.C("mail_bounces").Insert(payload)

		if err != nil {
			panic(err)
		}

		order, err := m.Store.OrderFinder(payload.Email)

		if err == nil {

			// Keep the track of what disabled the order
			order.PushTag("disabled:" + payload.Type)

			// Ignore the order
			order.Ignore()
		}

		c.JSON(200, gin.H{"status": "okay"})
	}

	c.JSON(400, gin.H{"message": "The request did not match the needed payload. Aborting..."})
}

func (m MailAPI) OpenWebhook(c *gin.Context) {

	var payload PostmarkOpenPayload

	if c.Bind(&payload) == nil {

		payload.Id = bson.NewObjectId()
		db := m.Mongo.Database
		err := db.C("mail_opens").Insert(payload)

		if err != nil {
			panic(err)
		}

		m.Store.TrackEmailOpened(payload.MessageID, payload.Id, payload.ReadSeconds)

		c.JSON(200, gin.H{"status": "okay"})
	}

	c.JSON(400, gin.H{"message": "The request did not match the needed payload. Aborting..."})
}

type PostmarkBouncedPayload struct {
	ID          int       `bson:"_id" json:"ID"`
	Type        string    `bson:"type" json:"Type"`
	TypeCode    int       `bson:"type_code" json:"TypeCode"`
	MessageID   string    `bson:"message_id" json:"MessageID"`
	Description string    `bson:"description" json:"Description"`
	Details     string    `bson:"details" json:"Details"`
	Inactive    bool      `bson:"inactive" json:"Inactive"`
	CanActivate bool      `bson:"can_activate" json:"CanActivate"`
	Email       string    `bson:"email" json:"Email"`
	Subject     string    `bson:"subject" json:"Subject"`
	BouncedAt   time.Time `bson:"bounced_at" json:"BouncedAt"`
}

type PostmarkOpenPayload struct {
	Id          bson.ObjectId          `bson:"_id,omitempty" json:"-"`
	MessageID   string                 `bson:"message_id" json:"MessageID"`
	Recipient   string                 `bson:"recipient" json:"Recipient"`
	ReadSeconds int                    `bson:"read_seconds" json:"ReadSeconds"`
	FirstOpen   bool                   `bson:"first_open" json:"FirstOpen"`
	Geo         map[string]interface{} `bson:"geo" json:"Geo"`
	ReceivedAt  time.Time              `bson:"received_at" json:"ReceivedAt"`
}

type MailInbound struct {
	Id                bson.ObjectId     `bson:"_id,omitempty" json:"id"`
	MessageID         string            `json:"MessageID"`
	From              string            `json:"From"`
	FromName          string            `json:"FromName"`
	FromFull          MailFullInbound   `json:"FromFull"`
	To                string            `json:"To"`
	ToFull            []MailFullInbound `json:"ToFull"`
	Cc                string            `json:"Cc"`
	CcFull            []MailFullInbound `json:"CcFull"`
	Bcc               string            `json:"Bcc"`
	BccFull           []MailFullInbound `json:"BccFull"`
	Subject           string            `json:"Subject"`
	Date              string            `json:"Date"`
	ReplyTo           string            `json:"ReplyTo"`
	MailboxHash       string            `json:"MailboxHash"`
	OriginalRecipient string            `json:"OriginalRecipient"`
	TextBody          string            `json:"TextBody"`
	HtmlBody          string            `json:"HtmlBody"`
	StrippedTextReply string            `json:"StrippedTextReply"`

	Attachments []MailAttachmentInbound `bson:"-" json:"Attachments"`
	Created     time.Time               `bson:"created_at,omitempty" json:"created_at,omitempty"`
}

type MailFullInbound struct {
	Email       string `json:"Email"`
	Name        string `json:"Name"`
	MailboxHash string `json:"MailboxHash"`
}

type MailAttachmentInbound struct {
	Name          string `json:"Name"`
	Content       string `json:"Content" bson:"-"`
	ContentType   string `json:"ContentType"`
	ContentLength int    `json:"ContentLength"`
	ContentID     string `json:"ContentID"`
}

type MailMandrillInbound struct {
	Id        bson.ObjectId              `bson:"_id,omitempty" json:"id,omitempty"`
	Timestamp int                        `json:"ts"`
	Event     string                     `json:"event"`
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
