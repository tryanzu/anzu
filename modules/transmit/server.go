package transmit

import (
	"github.com/googollee/go-socket.io"
	zmq "github.com/pebbe/zmq4"
	"github.com/rs/cors"

	"encoding/json"
	"log"
	"net/http"
	"sync"
)

func RunServer(deps Deps, socketPort, pullPort string) {
	var wg sync.WaitGroup
	wg.Add(2)

	redis := deps.Cache()
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

			// Async message saving.
			go func() {
				if message.Room == "chat" {
					if _, err := redis.LPush(message.RoomID(), msg); err != nil {
						log.Println("error:", err)
					}
				}
			}()

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

		server.On("connection", handleConnection(deps))
		server.On("error", func(so socketio.Socket, err error) {
			log.Println("error:", err)
		})

		go func() {
			for {
				msg := <-messages
				log.Printf("%v\n", msg)

				go server.BroadcastTo(msg.Room, msg.Event, msg.Message)
			}
		}()

		log.Println("Started sockets server at localhost:" + socketPort + "...")

		mux := http.NewServeMux()
		mux.Handle("/socket.io/", server)
		handler := cors.New(cors.Options{
			AllowCredentials: true,
		}).Handler(mux)

		log.Fatal(http.ListenAndServe(":"+socketPort, handler))
	}()

	log.Println("Waiting To Finish")
	wg.Wait()
}
