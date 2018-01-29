package realtime

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/desertbit/glue"
	"github.com/henrylee2cn/goutil"
	"github.com/tryanzu/core/deps"
)

var (
	jwtSecret  []byte
	server     *glue.Server
	sockets    goutil.Map
	clients    goutil.Map
	dispatcher chan []M

	// BufferSize holds the queue size for broadcasting channels.
	BufferSize = 10

	// Broadcast receives messages to be broadcasted into global namespace.
	Broadcast chan string

	// ToChan broadcasting
	ToChan chan M
)

// M holds the structure to broadcast messages
type M struct {
	Channel string
	Content string
}

type socketEvent struct {
	Event  string                 `json:"event"`
	Params map[string]interface{} `json:"params"`
}

func (ev socketEvent) encode() string {
	bytes, err := json.Marshal(ev)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func prepare() {
	sockets = goutil.AtomicMap()
	clients = goutil.RwMap(1000)

	// Prepare multicast channels before starting server
	Broadcast = make(chan string, BufferSize)
	ToChan = make(chan M, BufferSize)
	dispatcher = make(chan []M, BufferSize)

	// Bootstrap glue server instance
	server = glue.NewServer(glue.Options{
		HTTPSocketType: glue.HTTPSocketTypeNone,
	})

	secret, err := deps.Container.Config().String("application.secret")
	if err != nil {
		log.Panic("Could not get JWT secret token. (missing config)", err)
	}

	jwtSecret = []byte(secret)

	go func() {
		buffered := make([]M, 0, 1000)

		for {
			select {
			case m := <-Broadcast:
				buffered = append(buffered, M{Content: m})
			case m := <-ToChan:
				buffered = append(buffered, m)
			case <-time.After(time.Millisecond * 100):
				if len(buffered) > 0 {
					mark := elapsed("Flushing")
					log.Println("Flushing buffer with", len(buffered), "items.")
					dispatcher <- buffered
					buffered = make([]M, 0, 1000)
					mark()
				}
			}
		}
	}()

	go func() {
		for pack := range dispatcher {
			mark := elapsed("Dispatching")
			log.Printf("Messages: %+v\n", pack)
			sockets.Range(func(k, v interface{}) bool {
				c := v.(*Client)
				c.send(pack)
				return true
			})
			mark()
		}
	}()

	server.OnNewSocket(onNewSocket)
}

func onNewSocket(s *glue.Socket) {
	client := &Client{
		Raw:      s,
		Channels: make(map[string]*glue.Channel),
		User:     nil,
		Read:     make(chan socketEvent),
	}

	go client.readWorker()

	// Set a function which is triggered as soon as the socket is closed.
	s.OnClose(func() {
		sockets.Delete(s.ID())
		log.Printf("socket %s closed with remote address: %s", s.ID(), s.RemoteAddr())
	})

	// Set a function which is triggered during each received message.
	s.OnRead(func(data string) {
		var event socketEvent

		err := json.Unmarshal([]byte(data), &event)
		if err != nil {
			log.Println("Could not unmarshal read event from client: ", data)
			log.Println("Error: ", err)
		}

		client.Read <- event
	})

	// Send a welcome string to the client.
	s.Write(`{"event": "connected"}`)
	sockets.Store(s.ID(), client)
}

// ServeHTTP exposes http server handler for glue.
func ServeHTTP() func(w http.ResponseWriter, r *http.Request) {
	// Prepare server to handle requests.
	prepare()

	return server.ServeHTTP
}

func elapsed(name string) func() {
	starts := time.Now()
	return func() {
		ends := time.Now()
		elapsed := ends.Sub(starts)
		log.Printf("%s took %s", name, elapsed)
	}
}
