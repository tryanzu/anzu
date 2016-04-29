package store

import (
	"github.com/fernandez14/go-enlacefiscal"
	"github.com/fernandez14/spartangeek-blacker/modules/assets"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/mgo.v2/bson"

	"io/ioutil"
	"strconv"
	"strings"
	"time"
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
				Template: 250241,
				Recipient: []mail.MailRecipient{
					{
						Name:  self.data.User.Name,
						Email: self.data.User.Email,
					},
				},
				FromEmail: "pc@spartangeek.com",
				FromName:  "Drak Spartan",
				Variables: map[string]interface{}{
					"content": text,
				},
			}

			mailing.Send(compose)
		}()
	}
}

func (self *One) LoadDuplicates() {

	list := make([]OrderModel, 0)
	database := self.di.Mongo.Database

	where := []bson.M{{"user.email": self.data.User.Email}}

	if self.data.User.Ip != "" {
		where = append(where, bson.M{"user.ip": self.data.User.Ip})
	}

	err := database.C("orders").Find(bson.M{"$or": where}).All(&list)

	if err != nil {
		panic(err)
	}

	if len(list) > 0 {
		self.data.Duplicates = list
	}
}

func (self *One) LoadAssets() {

	var list []bson.ObjectId

	database := self.di.Mongo.Database

	for _, msg := range self.data.Messages {

		if !msg.RelatedId.Valid() {
			continue
		}

		if duplicated, _ := helpers.InArray(msg.RelatedId, list); !duplicated {

			list = append(list, msg.RelatedId)
		}
	}

	var als []assets.Asset

	err := database.C("assets").Find(bson.M{"related": "mail", "related_id": bson.M{"$in": list}}).All(&als)

	if err != nil {
		panic(err)
	}

	messages := self.data.Messages

	for key, msg := range messages {

		for _, asset := range als {

			if asset.RelatedId == msg.RelatedId {

				if l, init := messages[key].Meta["assets"]; init {

					a := l.([]assets.Asset)
					a = append(a, asset)

					messages[key].Meta["assets"] = a

				} else {

					var a []assets.Asset

					a = append(a, asset)

					messages[key].Meta = map[string]interface{}{
						"assets": a,
					}
				}
			}
		}
	}

	self.data.Messages = messages
}

func (self *One) LoadInvoice() {

	var invoice Invoice

	database := self.di.Mongo.Database
	err := database.C("invoices").Find(bson.M{"deal_id": self.data.Id}).One(&invoice)

	if err == nil {
		self.data.Invoice = &invoice
	}
}

func (self *One) PushTag(tag string) {

	database := self.di.Mongo.Database
	item := TagModel{
		Name:    tag,
		Created: time.Now(),
	}

	err := database.C("orders").Update(bson.M{"_id": self.data.Id}, bson.M{"$push": bson.M{"tags": item}})

	if err != nil {
		panic(err)
	}
}

func (self *One) PushActivity(name, description string, due_at time.Time) {

	database := self.di.Mongo.Database
	activity := ActivityModel{
		Name:        name,
		Description: description,
		Done:        false,
		Due:         due_at,
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	err := database.C("orders").Update(bson.M{"_id": self.data.Id}, bson.M{"$push": bson.M{"activities": activity}})

	if err != nil {
		panic(err)
	}
}

func (self *One) PushInboundAnswer(text string, mail bson.ObjectId) {

	database := self.di.Mongo.Database

	message := MessageModel{
		Content:   text,
		Type:      "inbound",
		RelatedId: mail,
		Created:   time.Now(),
		Updated:   time.Now(),
	}

	err := database.C("orders").Update(bson.M{"_id": self.data.Id}, bson.M{"$push": bson.M{"messages": message}, "$set": bson.M{"unreaded": true, "updated_at": time.Now()}})

	if err != nil {
		panic(err)
	}
}

func (self *One) Stage(name string) {

	// Temp way to validate the name of the stage
	if name != "estimate" && name != "negotiation" && name != "accepted" && name != "awaiting" && name != "closed" {
		return
	}

	database := self.di.Mongo.Database

	// Define steps in order
	steps := []string{"estimate", "negotiation", "accepted", "awaiting", "closed"}
	current := self.data.Pipeline.Step

	if current > 0 {
		current = current - 1
	}

	target := 0

	for index, step := range steps {

		if step == name {

			target = index
		}
	}

	named := steps[target]
	err := database.C("orders").Update(bson.M{"_id": self.data.Id}, bson.M{"$set": bson.M{"pipeline.step": target + 1, "pipeline.current": named, "pipeline.updated_at": time.Now(), "updated_at": time.Now()}})

	if err != nil {
		panic(err)
	}
}

func (self *One) MatchUsers() []user.UserBasic {

	database := self.di.Mongo.Database
	ip := self.data.User.Ip

	if ip != "" {

		var checkins []user.CheckinModel
		var users_id []bson.ObjectId

		err := database.C("checkins").Find(bson.M{"client_ip": ip}).All(&checkins)

		if err != nil {
			panic(err)
		}

		for _, checkin := range checkins {

			duplicated, _ := helpers.InArray(checkin.UserId, users_id)

			if !duplicated {

				users_id = append(users_id, checkin.UserId)
			}
		}

		var users []user.UserBasic

		err = database.C("users").Find(bson.M{"$or": []bson.M{
			{"_id": bson.M{"$in": users_id}},
			{"email": self.data.User.Email},
			{"facebook.email": self.data.User.Email},
		}}).Select(bson.M{"_id": 1, "username": 1, "username_slug": 1, "email": 1, "gaming": 1, "facebook": 1, "validated": 1, "banned": 1, "created_at": 1, "updated_at": 1}).All(&users)

		if err != nil {
			panic(err)
		}

		return users
	}

	var users []user.UserBasic

	err := database.C("users").Find(bson.M{"$or": []bson.M{
		{"email": self.data.User.Email},
		{"facebook.email": self.data.User.Email},
	}}).Select(bson.M{"_id": 1, "username": 1, "username_slug": 1, "email": 1, "facebook": 1, "validated": 1, "banned": 1, "created_at": 1, "updated_at": 1}).All(&users)

	if err != nil {
		panic(err)
	}

	return users
}

func (self *One) Touch() {

	database := self.di.Mongo.Database

	err := database.C("orders").Update(bson.M{"_id": self.data.Id}, bson.M{"$set": bson.M{"unreaded": false}})

	if err != nil {
		panic(err)
	}
}

func (self *One) Ignore() {

	database := self.di.Mongo.Database
	err := database.C("orders").Update(bson.M{"_id": self.data.Id}, bson.M{"$set": bson.M{"deleted_at": time.Now()}})

	if err != nil {
		panic(err)
	}
}

func (self *One) EmitInvoice(name, rfc, email string, total float64) (*Invoice, error) {

	config, err := self.di.Config.Get("invoicing")

	if err != nil {
		panic(err)
	}

	apiUser, err := config.String("username")

	if err != nil {
		panic(err)
	}

	apiPass, err := config.String("password")

	if err != nil {
		panic(err)
	}

	rfcOrigin, err := config.String("rfc")

	if err != nil {
		panic(err)
	}

	series, err := config.String("series")

	if err != nil {
		panic(err)
	}

	folioPath, err := config.String("folio")

	if err != nil {
		panic(err)
	}

	folioContent, err := ioutil.ReadFile(folioPath)

	if err != nil {
		panic(err)
	}

	folio, err := strconv.Atoi(string(folioContent))

	if err != nil {
		panic(err)
	}

	api := efiscal.Boot(apiUser, apiPass, true)
	invoice := api.Invoice(rfcOrigin, series, strconv.Itoa(folio))

	invoice.AddItem(efiscal.Item{
		Quantity:    1,
		Value:       total,
		Unit:        "pc",
		Description: "PC de Alto Rendimiento",
	})
	invoice.TransferIVA(16)
	invoice.SetPayment(&efiscal.PAY_ONE_TIME_TRANSFER)
	invoice.SetReceiver(&efiscal.Receiver{rfc, name, nil})
	invoice.SendMail([]string{email})

	data, err := api.Sign(invoice)

	var record *Invoice

	if err == nil {

		database := self.di.Mongo.Database
		record = &Invoice{
			Id:     bson.NewObjectId(),
			DealId: self.Data().Id,
			Assets: InvoiceAssets{
				PDF: "",
				XML: "",
			},
			Meta:    data,
			Created: time.Now(),
			Updated: time.Now(),
		}

		err := database.C("invoices").Insert(record)

		if err != nil {
			panic(err)
		}

		newFolio := strconv.Itoa(folio + 1)
		err = ioutil.WriteFile(folioPath, []byte(newFolio), 0644)

		if err != nil {
			return record, err
		}
	}

	return record, err
}
