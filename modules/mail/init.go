package mail

import (
	"github.com/hjr265/postmark.go/postmark"
	"github.com/olebedev/config"
	"gopkg.in/op/go-logging.v1"

	"bufio"
	"net/mail"
	"os"
	"strings"
)

type Module struct {
	Logger *logging.Logger `inject:""`
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
			if module.IsSafe(recipient.Email) {
				recipients = append(recipients, &mail.Address{
					Name:    recipient.Name,
					Address: recipient.Email,
				})
			}
		}
	}

	if len(recipients) == 0 {
		return "no-recipients"
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

func (m Module) IsSafe(email string) bool {
	for _, domain := range m.config.IgnoredDomains {
		if strings.HasSuffix(email, domain) {
			m.Logger.Warningf("%s has been declared as unsafe.", email)
			return false
		}
	}

	m.Logger.Infof("%s seems like a safe email.", email)
	return true
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

	ignoredPath, err := config.String("ignored")

	if err != nil {
		panic(err)
	}

	ignoredList := []string{}
	ignored, err := os.Open(ignoredPath)

	if err != nil {
		panic(err)
	}

	defer ignored.Close()

	scanner := bufio.NewScanner(ignored)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		ignoredList = append(ignoredList, scanner.Text())
	}

	module_config := ModuleConfig{
		From:           email,
		FromName:       name,
		Recipients:     recipients,
		IgnoredDomains: ignoredList,
	}

	module := &Module{debug: debug, config: module_config, Client: client}

	return module
}
