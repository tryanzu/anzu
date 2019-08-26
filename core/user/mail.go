package user

import (
	"errors"
	"time"

	"github.com/matcornic/hermes/v2"
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
	token, err := helpers.GenerateRandomStringURLSafe(32)
	if err != nil {
		return err
	}
	r := recoveryToken{
		UserID:  u.Id,
		Token:   token,
		Used:    false,
		Created: time.Now(),
		Updated: time.Now(),
	}
	err = d.Mgo().C("user_recovery_tokens").Insert(r)
	if err != nil {
		return err
	}
	c := config.C.Copy()
	m := gomail.NewMessage()
	from := c.Mail.From
	if len(from) == 0 {
		from = "no-reply@tryanzu.com"
	}
	h := config.C.Hermes()
	body, err := h.GenerateHTML(resetPasswordEmail(r, u))
	if err != nil {
		return err
	}
	m.SetHeader("From", from)
	m.SetHeader("Reply-To", from)
	m.SetHeader("To", u.Email)
	m.SetHeader("Subject", "Recupera tu acceso a "+c.Site.Name)
	m.SetBody("text/html", body)

	// Send email message.
	mail.In <- m

	return nil
}

func resetPasswordEmail(token recoveryToken, u User) hermes.Email {
	c := config.C.Copy()
	link := c.Site.MakeURL("recovery/" + token.Token)
	return hermes.Email{
		Body: hermes.Body{
			Name: u.UserName,
			Intros: []string{
				"You have received this email because a password reset request for your " + c.Site.Name + " account was received.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to reset your password:",
					Button: hermes.Button{
						Color: "#3D5AFE",
						Text:  "Reset your password",
						Link:  link,
					},
				},
			},
			Outros: []string{
				"If you did not request a password reset, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}
