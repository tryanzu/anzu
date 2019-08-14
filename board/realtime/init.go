package realtime

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/desertbit/glue"
	"github.com/henrylee2cn/goutil"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

var (
	jwtSecret  []byte
	server     *glue.Server
	sockets    *sync.Map
	addresses  *sync.Map
	clients    goutil.Map
	dispatcher chan []M
	counters   chan *Client

	// BufferSize holds the queue size for broadcasting channels.
	BufferSize = 100

	// Broadcast receives messages to be broadcasted into global namespace.
	Broadcast chan string

	// ToChan broadcasting
	ToChan chan M
)

// M holds the structure to broadcast messages
type M struct {
	Channel string
	Content string
	ID      *bson.ObjectId
}

type SocketEvent struct {
	Event  string                 `json:"event"`
	Params map[string]interface{} `json:"params"`
}

func (ev SocketEvent) encode() string {
	bytes, err := json.Marshal(ev)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func (ev SocketEvent) Encode() string {
	bytes, err := json.Marshal(ev)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func prepare() {
	sockets = new(sync.Map)
	addresses = new(sync.Map)
	clients = goutil.RwMap(1000)

	// Prepare multicast channels before starting server
	Broadcast = make(chan string, BufferSize)
	ToChan = make(chan M, BufferSize)
	featuredM = make(chan M)
	dispatcher = make(chan []M, BufferSize)
	counters = make(chan *Client, BufferSize)

	// Bootstrap glue server instance
	options := glue.Options{
		HTTPSocketType: glue.HTTPSocketTypeNone,
	}

	if env, err := deps.Container.Config().String("environment"); err == nil && env == "development" {
		options.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	server = glue.NewServer(options)
	secret, err := deps.Container.Config().String("application.secret")
	if err != nil {
		log.Panic("Could not get JWT secret token. (missing config)", err)
	}

	jwtSecret = []byte(secret)
	ledisdb := deps.Container.LedisDB()

	go func() {
		buffered := make([]M, 0, 1000)

		for {
			select {
			case m := <-Broadcast:
				buffered = append(buffered, M{Content: m})
			case m := <-ToChan:
				buffered = append(buffered, m)
			case <-time.After(time.Millisecond * 60):
				if len(buffered) == 0 {
					continue
				}
				mark := elapsed("Flushing")
				log.Println("[glue] Flushing buffer with", len(buffered), "items.")
				dispatcher <- buffered
				buffered = make([]M, 0, 1000)
				mark()
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
			for _, msg := range pack {
				if len(msg.Channel) > 5 && msg.Channel[0:5] == "chat:" {
					var (
						buf bytes.Buffer
						n   int64
					)
					enc := gob.NewEncoder(&buf)
					err := enc.Encode(msg)
					if err != nil {
						log.Println("[glue] [err] Cannot encode for cache", err)
					}
					n, err = ledisdb.RPush([]byte(msg.Channel), buf.Bytes())
					if err != nil {
						log.Println("[glue] [err] Cannot encode for cache", err)
					}
					if n >= 50 {
						ledisdb.LPop([]byte(msg.Channel))
					}
				}
			}
			mark()
		}
	}()

	go countClientsWorker()
	go starredMessagesWorker()

	server.OnNewSocket(onNewSocket)
}

func onNewSocket(s *glue.Socket) {
	addr := s.RemoteAddr()
	if n, ok := addresses.LoadOrStore(addr, 1); ok {
		actual := n.(int) + 1
		if actual > 3 {
			return
		}
		addresses.Store(addr, actual)
	}
	client := &Client{
		Raw:      s,
		Channels: new(sync.Map),
		User:     nil,
		Read:     make(chan SocketEvent),
	}

	go client.readWorker()

	// Set a function which is triggered as soon as the socket is closed.
	s.OnClose(func() {
		close(client.Read)
		addresses.Delete(client.Raw.RemoteAddr())
		client.User = nil
		client.Channels = nil
		client.Raw = nil
		sockets.Delete(s.ID())
		counters <- client
		log.Printf("[glue] Socket %s closed with remote address: %s", s.ID(), s.RemoteAddr())
	})

	// fn triggered during each received message.
	s.OnRead(func(data string) {
		var event SocketEvent

		err := json.Unmarshal([]byte(data), &event)
		if err != nil {
			log.Println("Could not unmarshal read event from client: ", data)
			log.Println("Error: ", err)
			return
		}

		client.Read <- event
	})

	runtime := config.C.Copy()
	conf, err := json.Marshal(map[string]interface{}{
		"event":  "config",
		"params": runtime.Site,
	})

	if err != nil {
		panic(err)
	}

	// Send a welcome string to the client.
	s.Write(`{"event": "connected"}`)
	s.Write(string(conf))

	sockets.Store(s.ID(), client)
}

// ServeHTTP exposes http server handler for glue.
func ServeHTTP() func(w http.ResponseWriter, r *http.Request) {
	// Prepare server to handle requests.
	prepare()

	go func() {
		for {
			<-config.C.Reload
			runtime := config.C.Copy()
			conf, err := json.Marshal(map[string]interface{}{
				"event":  "config",
				"params": runtime.Site,
			})
			if err != nil {
				panic(err)
			}

			Broadcast <- string(conf)
		}
	}()

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
