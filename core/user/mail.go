package user

import (
	"errors"
	"fmt"
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

func (u User) EnforceAccountValidationEmail(d deps) (err error) {
	c := config.C.Copy()
	m := gomail.NewMessage()
	from := c.Mail.From
	if len(from) == 0 {
		from = "no-reply@tryanzu.com"
	}
	h := config.C.Hermes()
	body, err := h.GenerateHTML(requestAccountValidation(u))
	if err != nil {
		return err
	}
	m.SetHeader("From", from)
	m.SetHeader("Reply-To", from)
	m.SetHeader("To", u.Email)
	m.SetHeader("Subject", "Verifica el correo de tu cuenta en "+c.Site.Name)
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

func requestAccountValidation(u User) hermes.Email {
	c := config.C.Copy()
	link := c.Site.MakeURL("validate/" + u.VerificationCode)
	return hermes.Email{
		Body: hermes.Body{
			Name: u.UserName,
			Intros: []string{
				fmt.Sprintf("Tu cuenta en %s (%s) aún no ha sido validada, si recuerdas tu registro en %s ayudanos a validarla a la brevedad.", c.Site.Name, u.UserName, c.Site.Name),
				fmt.Sprintf("Si tu cuenta no es validada en las proximas 24 horas, borraremos esta cuenta y sus datos de nuestra plataforma."),
			},
			Actions: []hermes.Action{
				{
					Instructions: "Da click en el boton para validar el correo electrónico de tu cuenta:",
					Button: hermes.Button{
						Color: "#3D5AFE",
						Text:  "Validar mi cuenta",
						Link:  link,
					},
				},
			},
			Outros: []string{
				fmt.Sprintf("Si no recuerdas tu registro en %s o no te interesa mantener una cuenta en nuestra comunidad, ignora este mensaje.", c.Site.Name),
			},
			Signature: "Gracias",
		},
	}
}
