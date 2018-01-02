package realtime

import (
	"log"
	"net/http"
	"time"

	"github.com/desertbit/glue"
	"github.com/henrylee2cn/goutil"
	"github.com/tryanzu/core/core/user"
)

var (
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

// Client contains a message to be broadcasted to a channel
type Client struct {
	Raw      *glue.Socket
	Channels map[string]*glue.Channel
	User     *user.User
	Read     chan<- readEvent
}

type readEvent struct {
	Event  string
	Params map[string]interface{}
}

func (c *Client) send(packed []M) {
	for _, m := range packed {
		if m.Channel == "" {
			c.Raw.Write(m.Content)
			continue
		}

		if _, exists := c.Channels[m.Channel]; exists == false {
			c.Channels[m.Channel] = c.Raw.Channel(m.Channel)
		}

		c.Channels[m.Channel].Write(m.Content)
	}
}

func init() {
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

	go func() {
		buffered := make([]M, 0, 1000)

		for {
			select {
			case m := <-Broadcast:
				buffered = append(buffered, M{Content: m})
			case m := <-ToChan:
				buffered = append(buffered, m)
			case <-time.After(time.Second):
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
		Read:     make(chan readEvent),
	}

	// Set a function which is triggered as soon as the socket is closed.
	s.OnClose(func() {
		sockets.Delete(s.ID())
		log.Printf("socket %s closed with remote address: %s", s.ID(), s.RemoteAddr())
	})

	// Set a function which is triggered during each received message.
	s.OnRead(func(data string) {

		// Echo the received data back to the client.
		s.Close()
	})

	// Send a welcome string to the client.
	s.Write(`{"event": "connected"}`)
	sockets.Store(s.ID(), client)
}

// ServeHTTP exposes http server handler for glue.
func ServeHTTP() func(w http.ResponseWriter, r *http.Request) {
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
