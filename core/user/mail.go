package user

import (
	"errors"
	"time"

	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
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
	compose := mail.Mail{
		mail.MailBase{
			Recipient: []mail.MailRecipient{
				{
					Name:  u.UserName,
					Email: u.Email,
				},
			},
			Variables: map[string]interface{}{
				"confirm_url": "https://spartangeek.com/signup/confirm/" + u.VerificationCode,
			},
		},
		250222,
	}

	update := bson.M{"$set": bson.M{"confirm_sent_at": time.Now()}}
	err = users.Update(bson.M{"_id": u.Id}, update)
	if err != nil {
		return
	}

	d.Mailer().Send(compose)
	return
}

func (u User) RecoveryPasswordEmail(d Deps) (err error) {
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

	compose := mail.Mail{
		mail.MailBase{
			Recipient: []mail.MailRecipient{
				{
					Name:  u.UserName,
					Email: u.Email,
				},
			},
			Variables: map[string]interface{}{
				"recover_url": "https://spartangeek.com/user/lost_password/" + r.Token,
			},
		},
		461461,
	}

	d.Mailer().Send(compose)
	return
}
