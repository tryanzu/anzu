package mail

import (
	"log"
	"time"

	"github.com/tryanzu/core/core/config"
	gomail "gopkg.in/gomail.v2"
)

var (
	// In channel will receive the messages to be sent.
	In chan *gomail.Message
)

// Boot send worker with automatic runtime config.
func Boot() {
	go func() {
		for {
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
		log.Println("[BOOT] Mail send worker has started...", mail)
		dialer := gomail.NewPlainDialer(mail.Server, 587, mail.User, mail.Password)
		for {
			select {
			case m, alive := <-In:
				if !alive {
					log.Println("Mail send worker has stopped...")
					return
				}
				if !open {
					if sender, err = dialer.Dial(); err != nil {
						panic(err)
					}
					open = true
				}
				if err := gomail.Send(sender, m); err != nil {
					log.Print(err)
				}
			case <-time.After(30 * time.Second):
				// Close the connection to the SMTP server if no email was sent in
				// the last 30 seconds.
				if open {
					if err := sender.Close(); err != nil {
						panic(err)
					}
					open = false
				}
			}
		}
	}

	log.Println("[BOOT] Mail worker is not configured, discarding emails...")
	for {
		m, alive := <-In
		if !alive {
			log.Println("[BOOT] Mail bad worker has stopped...")
			return
		}

		log.Println("[BOOT] Discarding mail message", m)
	}
}
