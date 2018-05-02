package user

import (
	"errors"
	"time"

	"github.com/tryanzu/core/core/mail"
	"github.com/tryanzu/core/core/templates"
	"github.com/tryanzu/core/modules/helpers"
	gomail "gopkg.in/gomail.v2"
	"gopkg.in/mgo.v2/bson"
)

func (u User) ConfirmationEmail(d Deps) (err error) {

	// Check last confirmation email sent.
	if u.ConfirmSent != nil {
		deadline := u.ConfirmSent.Add(time.Duration(time.Minute * 5))
		if deadline.After(time.Now()) {
			err = errors.New("Rate exceeded temporarily")
			return
		}
	}

	users := d.Mgo().C("users")
	body, err := templates.ExecuteTemplate("mails/welcome", map[string]interface{}{
		"user": u,
	})
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "contacto@spartangeek.com")
	m.SetHeader("Reply-To", "contacto@spartangeek.com")
	m.SetHeader("To", u.Email)
	m.SetHeader("Subject", "Bienvenido a SpartanGeek")
	m.SetBody("text/html", body.String())

	mail.In <- m

	update := bson.M{"$set": bson.M{"confirm_sent_at": time.Now()}}
	err = users.Update(bson.M{"_id": u.Id}, update)
	if err != nil {
		return
	}

	return
}

func (u User) RecoveryPasswordEmail(d Deps) (err error) {
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
