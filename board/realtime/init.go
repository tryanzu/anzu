package realtime

import (
	"log"
	"net/http"
	"sync"

	"github.com/desertbit/glue"
)

// Input channel buffer size.
const BufferSize = 10

var (
	server  *glue.Server
	sockets *sync.Map

	// Public accesible
	Broadcast chan string
)

func init() {
	sockets = &sync.Map{}

	// Prepare multicast channels before starting server
	Broadcast = make(chan string, BufferSize)

	// Bootstrap glue server instance
	server = glue.NewServer(glue.Options{
		HTTPSocketType: glue.HTTPSocketTypeNone,
	})

	go func() {
		for m := range Broadcast {
			go func(m string) {
				log.Println("Broadcasting ", m)
				sockets.Range(func(k, v interface{}) bool {
					s := v.(*glue.Socket)
					s.Write(m)
					return true
				})
			}(m)
		}
	}()

	server.OnNewSocket(onNewSocket)
}

func onNewSocket(s *glue.Socket) {
	// Set a function which is triggered as soon as the socket is closed.
	s.OnClose(func() {
		sockets.Delete(s.ID())
		log.Printf("socket %s closed with remote address: %s", s.ID(), s.RemoteAddr())
	})

	// Set a function which is triggered during each received message.
	s.OnRead(func(data string) {
		// Echo the received data back to the client.
		s.Write(data)
	})

	// Send a welcome string to the client.
	s.Write(`{"event": "connected"}`)
	sockets.Store(s.ID(), s)
}

func ServeHTTP() func(w http.ResponseWriter, r *http.Request) {
	return server.ServeHTTP
}
