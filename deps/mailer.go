package deps

import (
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
)

func IgniteMailer(container Deps) (Deps, error) {
	config, err := container.Config().Get("mail")
	if err != nil {
		return container, err
	}

	apiKey, err := container.Config().String("mail.api_key")
	if err != nil {
		return container, err
	}

	container.MailerProvider = mail.Boot(apiKey, config, false)
	return container, nil
}
