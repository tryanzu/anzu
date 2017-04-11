package transmit

import (
	"github.com/googollee/go-socket.io"
	zmq "github.com/pebbe/zmq4"
	"github.com/rs/cors"
	"github.com/xuyu/goredis"

	"encoding/json"
	"log"
	"net/http"
	"sync"
)

func RunServer(socketPort, pullPort string, redis *goredis.Redis) {
	var wg sync.WaitGroup
	wg.Add(2)

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
			if message.Room == "chat" {
				if _, err := redis.LPush(message.RoomID(), msg); err != nil {
					log.Println("error:", err)
				}
			}

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

		server.On("connection", func(so socketio.Socket) {
			log.Println("Connection handled.")
			so.Join("feed")
			so.Join("post")
			so.Join("general")
			so.Join("chat")
			so.Join("user")

			so.On("disconnection", func() {
				log.Println("Diconnection handled.")
			})

			so.On("chat update-me", func(channel string) {
				redis.LTrim("chat:"+channel, 0, 9)
				last, err := redis.LRange("chat:"+channel, 0, 9)
				if err == nil {
					for i := len(last) - 1; i >= 0; i-- {
						var m Message
						if err := json.Unmarshal([]byte(last[i]), &m); err != nil {
							continue
						}

						so.Emit("chat "+channel, m.Message)
					}
				}
			})
		})

		server.On("error", func(so socketio.Socket, err error) {
			log.Println("error:", err)
		})

		go func() {
			for {
				msg := <-messages
				log.Printf("%v\n", msg)

				server.BroadcastTo(msg.Room, msg.Event, msg.Message)
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
