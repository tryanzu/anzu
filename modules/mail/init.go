package mail

import (
	"github.com/hjr265/postmark.go/postmark"
	"github.com/olebedev/config"

	"net/mail"
)

type Module struct {
	Client *postmark.Client
	config ModuleConfig
	debug  bool
}

func (module Module) Send(m Mail) string {

	message := &postmark.Message{}

	if m.FromName == "" && m.FromEmail == "" {

		message.From = &mail.Address{
			Name:    module.config.FromName,
			Address: module.config.From,
		}

	} else {

		message.From = &mail.Address{
			Name:    m.FromName,
			Address: m.FromEmail,
		}
	}

	var recipients []*mail.Address

	if module.debug {

		for _, recipient := range module.config.Recipients {

			recipients = append(recipients, &mail.Address{
				Name:    recipient,
				Address: recipient,
			})
		}

	} else {

		for _, recipient := range m.Recipient {

			recipients = append(recipients, &mail.Address{
				Name:    recipient.Name,
				Address: recipient.Email,
			})
		}
	}

	message.To = recipients
	message.TemplateModel = m.Variables
	message.TemplateId = m.Template

	// Send the email using mandrill's API abstraction
	res, err := module.Client.Send(message)

	if err != nil {
		panic(err)
	}

	return res.MessageID
}

func Boot(key string, config *config.Config, debug bool) *Module {

	// Initialize mandrill client
	client := &postmark.Client{
		ApiKey: key,
		Secure: true,
	}

	name, err := config.String("from.name")

	if err != nil {
		panic(err)
	}

	email, err := config.String("from.email")

	if err != nil {
		panic(err)
	}

	list, err := config.List("recipients")

	if err != nil {
		panic(err)
	}

	recipients := make([]string, len(list)-1)

	for _, recipient := range list {

		recipients = append(recipients, recipient.(string))
	}

	module_config := ModuleConfig{
		From:       email,
		FromName:   name,
		Recipients: recipients,
	}

	module := &Module{debug: debug, config: module_config, Client: client}

	return module
}
