package notifications

import (
	"fmt"

	"github.com/matcornic/hermes/v2"
	"github.com/tryanzu/core/board/comments"
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

// SomeoneCommentedYourCommentEmail notification constructor.
func SomeoneCommentedYourCommentEmail(p post.Post, comment comments.Comment, usr user.User) (*gomail.Message, error) {
	c := config.C.Copy()
	link := c.Site.MakeURL("p/" + p.Slug + "/" + p.Id.Hex())
	email := hermes.Email{
		Body: hermes.Body{
			Name: usr.UserName,
			Intros: []string{
				fmt.Sprintf("Tu comentario en %s (Publicación %s) recibió una respuesta mientras no estabas.", c.Site.Name, p.Title),
			},
			Actions: []hermes.Action{
				{
					Button: hermes.Button{
						Color: "#3D5AFE",
						Text:  "Ver comentario",
						Link:  link,
					},
				},
			},
			Outros: []string{
				"Si deseas dejar de recibir notificaciones puedes entrar en tu cuenta y cambiar la configuración de avisos.",
			},
			Signature: "Un saludo",
		},
	}
	m := gomail.NewMessage()
	from := c.Mail.From
	if len(from) == 0 {
		from = "no-reply@tryanzu.com"
	}
	h := config.C.Hermes()
	body, err := h.GenerateHTML(email)
	if err != nil {
		return nil, err
	}
	m.SetHeader("From", from)
	m.SetHeader("Reply-To", from)
	m.SetHeader("To", usr.Email)
	m.SetHeader("Subject", "Alguien respondió tu comentario en: "+p.Title)
	m.SetBody("text/html", body)
	return m, nil
}
