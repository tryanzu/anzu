package notifications

import (
	post "github.com/tryanzu/core/board/posts"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/core/user"
	"gopkg.in/gomail.v2"
)

// SomeoneCommentedYourPostEmail notification constructor.
func SomeoneCommentedYourPostEmail(p post.Post, usr user.User) (*gomail.Message, error) {
	c := config.C.Copy()
	m := gomail.NewMessage()
	from := c.Mail.From
	if len(from) == 0 {
		from = "no-reply@tryanzu.com"
	}
	h := config.C.Hermes()
	body, err := h.GenerateHTML(post.SomeoneCommentedYourPost(usr.UserName, p))
	if err != nil {
		return nil, err
	}
	m.SetHeader("From", from)
	m.SetHeader("Reply-To", from)
	m.SetHeader("To", usr.Email)
	m.SetHeader("Subject", "Alguien respondió tu publicación: "+p.Title)
	m.SetBody("text/html", body)
	return m, nil
}
