package user

import (
	"errors"
	"time"

	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/mail"
	"github.com/tryanzu/core/core/templates"
	"github.com/tryanzu/core/modules/helpers"
	gomail "gopkg.in/gomail.v2"
	"gopkg.in/mgo.v2/bson"
)

func (u User) ConfirmationEmail(d deps) (err error) {
	// Check last confirmation email sent.
	if u.ConfirmSent != nil {
		deadline := u.ConfirmSent.Add(time.Duration(time.Minute * 5))
		if deadline.After(time.Now()) {
			err = errors.New("Confirmation email rate exceeded. Wait for another 5 minutes to try again")
			return
		}
	}
	body, err := templates.ExecuteTemplate("mails/welcome", map[string]interface{}{
		"user": u,
	})
	if err != nil {
		return err
	}

	c := config.C.Copy()
	m := gomail.NewMessage()
	from := c.Mail.From
	if len(from) == 0 {
		from = "no-reply@tryanzu.com"
	}

	m.SetHeader("From", from)
	m.SetHeader("Reply-To", from)
	m.SetHeader("To", u.Email)
	m.SetHeader("Subject", "Bienvenido a "+c.Site.Name)
	m.SetBody("text/html", body.String())

	// Send email message.
	mail.In <- m

	// Update sent at date.
	update := bson.M{"$set": bson.M{"confirm_sent_at": time.Now()}}
	err = d.Mgo().C("users").Update(bson.M{"_id": u.Id}, update)
	if err != nil {
		return
	}

	return
}

func (u User) RecoveryPasswordEmail(d deps) (err error) {
	return nil
	r := RecoveryToken{
		UserId:  u.Id,
		Token:   helpers.StrRandom(12),
		Used:    false,
		Created: time.Now(),
		Updated: time.Now(),
	}

	err = d.Mgo().C("user_recovery_tokens").Insert(r)
	if err != nil {
		return
	}

	return
}
