package mail

import (
	"github.com/keighl/mandrill"
	"github.com/olebedev/config"
)

type Module struct {
	Client     *mandrill.Client
	config     ModuleConfig
	debug      bool	
}

func (module Module) Send(mail Mail) {

	message := &mandrill.Message{}

	// Setup mail message
	message.FromEmail = module.config.From
	message.FromName = module.config.FromName

	if module.debug {

		for _, recipient := range module.config.Recipients {

			message.AddRecipient(recipient, recipient, "to")
		}

	} else {

		for _, recipient := range mail.Recipient {

			message.AddRecipient(recipient.Email, recipient.Name, "to")
		}
	}

	message.Subject = mail.Subject

	// Send the email using mandrill's API abstraction
	_, err := module.Client.MessagesSendTemplate(message, mail.Template, mail.Variables)

	if err != nil {
		panic(err)
	}
}

func Boot(key string, config *config.Config, debug bool) *Module {

	// Initialize mandrill client
	client := mandrill.ClientWithKey(key)

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

	recipients := make([]string, len(list))

	for _, recipient := range list {

		recipients = append(recipients, recipient.(string))
	}

	module_config := ModuleConfig{
		From: email,
		FromName: name, 
		Recipients: recipients,
	}

	module := &Module{debug: debug, config: module_config, Client: client}

	return module
}