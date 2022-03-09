package mail

import (
	"time"

	"github.com/op/go-logging"
	"github.com/tryanzu/core/core/config"
	gomail "gopkg.in/gomail.v2"
)

var (
	log = logging.MustGetLogger("mailer")

	// In channel will receive the messages to be sent.
	In chan *gomail.Message
)

// Boot send worker with automatic runtime config.
func Boot() {
	go func() {
		for {
			log.SetBackend(config.LoggingBackend)

			// New incoming messages chan && sendingWorkers
			In = make(chan *gomail.Message, 4)

			// Spawn a daemon that will consume incoming messages to be sent.
			// If not properly configured it will start ignoring signals.
			go sendWorker(config.C)

			<-config.C.Reload

			// When receiving the config signal
			// the In chan must be closed so active sendWorker
			// can finish.
			close(In)
		}
	}()
}

func sendWorker(c *config.Config) {
	var (
		sender gomail.SendCloser
		err    error
		open   = false
	)

	mail := c.Copy().Mail
	if len(mail.Server) > 0 {
		log.Info("send worker has started...", mail)
		dialer := gomail.NewDialer(mail.Server, 587, mail.User, mail.Password)
		for {
			select {
			case m, alive := <-In:
				if !alive {
					log.Info("Mail send worker has stopped...")
					return
				}
				if !open {
					if sender, err = dialer.Dial(); err != nil {
						log.Error(err)
						continue
					}
					open = true
				}
				if err := gomail.Send(sender, m); err != nil {
					log.Error(err)
					open = false
				}
			case <-time.After(30 * time.Second):
				// Close the connection to the SMTP server if no email was sent in
				// the last 30 seconds.
				if open {
					if err := sender.Close(); err != nil {
						log.Error(err)
						break
					}
					open = false
				}
			}
		}
	}

	log.Warning("mail settings are not configured, discarding emails...")
	for {
		m, alive := <-In
		if !alive {
			log.Info("worker has stopped...")
			return
		}

		log.Debug(m)
	}
}
