package transmit

import (
	"github.com/googollee/go-socket.io"
	zmq "github.com/pebbe/zmq4"

	"encoding/json"
	"log"
	"net/http"
	"sync"
)

func RunServer(deps Deps, socketPort, pullPort string) {
	var wg sync.WaitGroup
	wg.Add(2)

	//redis := deps.Cache()
	messages := make(chan Message, 100)

	go func() {
		defer wg.Done()

		//  Socket to receive messages on
		receiver, err := zmq.NewSocket(zmq.PULL)
		if err != nil {
			panic(err)
		}

		err = receiver.Bind("tcp://*:" + pullPort)
		if err != nil {
			panic(err)
		}

		log.Println("Started zmq pull server at tcp://*:" + pullPort)

		for {
			var message Message

			// Block until new message is received.
			msg, _ := receiver.Recv(0)

			if err := json.Unmarshal([]byte(msg), &message); err != nil {
				continue
			}

			// Once message is unmarshaled send it back to processing channel
			messages <- message

			log.Println("Broadcasted message to " + message.Room)
		}

		log.Println("Closed receiver")
		receiver.Close()
	}()

	go func() {
		defer wg.Done()

		server, err := socketio.NewServer(nil)
		if err != nil {
			log.Fatal(err)
		}

		handler := handleConnection(deps)

		server.On("connection", handler)
		server.On("error", func(so socketio.Socket, err error) {
			log.Println("error:", err)
		})

		go func() {
			for msg := range messages {
				go server.BroadcastTo(msg.Room, msg.Event, msg.Message)
			}
		}()

		log.Println("Started sockets server at localhost:" + socketPort + "...")

		http.Handle("/socket.io/", server)

		log.Fatal(http.ListenAndServe(":"+socketPort, nil))
	}()

	log.Println("Waiting To Finish")
	wg.Wait()
}
