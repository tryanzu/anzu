package queue

import (
	"encoding/json"
	"github.com/iron-io/iron_go/mq"
	"github.com/facebookgo/inject"
	"log"
	"os"
	"fmt"
)

// Handlers specification
type fn func (map[string]interface{})

type Module struct {
	Mail     MailJob
	Handlers map[string]fn
}

func (module *Module) Populate(g inject.Graph) {

	err := g.Provide(
		&inject.Object{Value: &module.Mail},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Populate the DI with the instances
	if err := g.Populate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	module.Handlers = map[string]fn {
		"StoreDelayedResponse": module.Mail.StoreDelayedResponse,
	}
}

func (module *Module) Listen(name string) {
	
	log.Printf("[info] Started worker for %v\n", name)

	queue := mq.New(name);

	for {

		// Get ten messages with a 60 seconds release time and wait 30 seconds in case of boring time
		msgs, err := queue.GetNWithTimeoutAndWait(10, 60, 30)

		if err != nil {

			log.Println(err.Error())
			continue
		}

		for _, msg := range msgs {

			module.doMessage(msg)

			// After firing message delete it from queue
			msg.Delete()
		}
	}
}

func (module *Module) doMessage(msg *mq.Message) {

	var message map[string]interface{}

	// Decode the message which must be a json message
	if err := json.Unmarshal([]byte(msg.Body), &message); err != nil {
        panic(err)
    }

    if _, handle := message["fire"]; handle {

    	// Once we've got the message we need to check if theres a handler for it
    	handle_name := message["fire"].(string)

    	if handler, handler_valid := module.Handlers[handle_name]; handler_valid {

    		// Execute the handler and pass the message as a parameter for it
    		handler(message)

    		return
    	}

    	panic("Invalid handler for the message")
    }

    panic("Not a valid message")
}