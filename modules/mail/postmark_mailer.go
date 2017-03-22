package mail

import (
	"github.com/hjr265/postmark.go/postmark"
	"github.com/olebedev/config"
	"github.com/op/go-logging"

	"bufio"
	"errors"
	"net/mail"
	"os"
	"strings"
)

var ErrInvalidRecipients = errors.New("Invalid recipients.")

type PostmarkMailer struct {
	Logger *logging.Logger
	Client *postmark.Client
	config ModuleConfig
	debug  bool
}

func (module PostmarkMailer) Send(m Mail) string {
	message, err := module.prepareMessage(m)
	if err != nil {
		panic(err)
	}

	message.TemplateModel = m.Variables
	message.TemplateId = m.Template

	// Send the email using mandrill's API abstraction
	res, err := module.Client.Send(message)

	if err != nil {
		panic(err)
	}

	return res.MessageID
}

func (module PostmarkMailer) SendRaw(raw Raw) string {
	message, err := module.prepareMessage(raw)
	if err != nil {
		panic(err)
	}

	message.HtmlBody = raw.Content

	// Send the email using mandrill's API abstraction
	res, err := module.Client.Send(message)

	if err != nil {
		panic(err)
	}

	return res.MessageID
}

func (m PostmarkMailer) IsSafe(email string) bool {
	for _, domain := range m.config.IgnoredDomains {
		if strings.HasSuffix(email, domain) {
			m.Logger.Warningf("%s has been declared as unsafe.", email)
			return false
		}
	}

	m.Logger.Debugf("%s seems like a safe email.", email)
	return true
}

func (module PostmarkMailer) prepareRecipients(m Mailable) []*mail.Address {
	if module.debug {
		var recipients []*mail.Address
		for _, recipient := range module.config.Recipients {
			recipients = append(recipients, &mail.Address{
				Name:    recipient,
				Address: recipient,
			})
		}

		return recipients
	}

	to := m.To()

	// Filter without allocating.
	recipients := to[:0]
	for _, res := range to {
		if module.IsSafe(res.Address) {
			recipients = append(recipients, res)
		}
	}

	return recipients
}

func (module PostmarkMailer) prepareMessage(m Mailable) (*postmark.Message, error) {
	message := &postmark.Message{}
	message.From = m.From()
	if message.From.Address == "" && message.From.Address == "" {
		message.From = &mail.Address{
			Name:    module.config.FromName,
			Address: module.config.From,
		}
	}

	recipients := module.prepareRecipients(m)
	if len(recipients) == 0 {
		return nil, ErrInvalidRecipients
	}

	message.To = recipients
	message.Subject = m.SubjectText()
	return message, nil
}

func Postmark(config *config.Config, logger *logging.Logger) (PostmarkMailer, error) {

	apiKey, err := config.String("api_key")
	if err != nil {
		return PostmarkMailer{}, err
	}

	// Initialize mandrill client
	client := &postmark.Client{
		ApiKey: apiKey,
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

	module := PostmarkMailer{config: module_config, Client: client, Logger: logger}

	return module, nil
}
