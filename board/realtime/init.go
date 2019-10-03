package realtime

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/desertbit/glue"
	"github.com/henrylee2cn/goutil"
	"github.com/op/go-logging"
	"github.com/tryanzu/core/core/config"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

var log = logging.MustGetLogger("realtime")

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
				mark := elapsed("flushing")
				log.Debugf("flushing buffer, items = %v", len(buffered))
				dispatcher <- buffered
				buffered = make([]M, 0, 1000)
				mark()
			}
		}
	}()

	go func() {
		for pack := range dispatcher {
			mark := elapsed("dispatching")
			log.Debugf("dispatching messages, list = %+v", pack)
			sockets.Range(func(k, v interface{}) bool {
				c := v.(*Client)
				c.send(pack)
				return true
			})
			mark()
		}
	}()

	go countClientsWorker()
	go starredMessagesWorker()

	server.OnNewSocket(onNewSocket)
}

func onNewSocket(s *glue.Socket) {
	addr := s.RemoteAddr()
	conns := 1
	if n, ok := addresses.LoadOrStore(addr, 1); ok {
		conns = n.(int) + 1
		if conns > 3 {
			return
		}
		addresses.Store(addr, conns)
	}
	client := &Client{
		Raw:      s,
		Channels: new(sync.Map),
		User:     nil,
		Read:     make(chan SocketEvent),
	}

	log.Infof("client connected, id = %s | address = %s | connections = %v", s.ID(), addr, conns)

	// READ WORKER
	// This little dedicated goroutine will handle all incoming messages for this particular client.
	go client.readWorker()

	// Set a function which is triggered as soon as the socket is closed.
	s.OnClose(client.finish)

	// fn triggered during each received message.
	s.OnRead(func(data string) {
		var event SocketEvent

		err := json.Unmarshal([]byte(data), &event)
		if err != nil {
			log.Errorf("Could not unmarshal read event from client: %s", data)
			log.Error(err)
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
		log.Debugf("time elapsed in, name = %s | took = %s", name, elapsed)
	}
}
