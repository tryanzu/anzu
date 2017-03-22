package deps

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
)

// Bootstraps mailer driver.
func IgniteMailer(container Deps) (Deps, error) {
	mailer := container.Config().UString("mail.driver", "test")
	config, err := container.Config().Get("mail")
	if err != nil {
		return container, err
	}

	switch mailer {
	case "test":
		container.MailerProvider = mail.TestMailer{Logger: container.Log()}
	case "postmark":
		container.MailerProvider, err = mail.Postmark(config, container.Log())
		if err != nil {
			return container, err
		}
	}

	return container, nil
}
