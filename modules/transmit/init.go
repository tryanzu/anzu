package transmit

import (
	"github.com/olebedev/config"
	"github.com/googollee/go-socket.io"
	zmq "github.com/pebbe/zmq4"

    "log"
    "net/http"
    "encoding/json"
)

type Module struct {
	Config *config.Config `inject:""`
}

func (module *Module) Run() {

	socketPort, err := module.Config.String("ports.socket")

	if err != nil {
		log.Fatal("Socket port not found in config.")
	}

	pullPort, err := module.Config.String("zmq.pull")

	if err != nil {
		log.Fatal("ZMQ pull port not found in config.")
	}

	server, err := socketio.NewServer(nil)
    
    if err != nil {
        log.Fatal(err)
    }

    server.SetAllowRequest(func(r *http.Request) error {

        origin := r.Header.Get("Origin")

        r.Header.Set("Access-Control-Allow-Origin", origin)

        return nil
    })
    
    server.On("connection", func(so socketio.Socket) {

        log.Println("Connection handled.")

        so.Join("feed")
        so.Join("post")
        so.Join("general")
        so.Join("user")
        
        so.On("disconnection", func() {
            log.Println("Diconnection handled.")
        })
    })

    server.On("error", func(so socketio.Socket, err error) {
        log.Println("error:", err)
    })

    go func() {

        //  Socket to receive messages on
        receiver, _ := zmq.NewSocket(zmq.PULL)
        defer receiver.Close()
        receiver.Connect("tcp://localhost:" + pullPort)

        for {

            var message Message 
            
            msg, _ := receiver.Recv(0)

            if err := json.Unmarshal([]byte(msg), &message); err != nil {
                continue
            }

            server.BroadcastTo(message.Room, message.Event, message.Message)
            log.Println("Broadcasted message to " + message.Room)
        }
    }()

    log.Println("Started sockets server at localhost:" + socketPort + "...")
    http.Handle("/socket.io/", server)
    log.Fatal(http.ListenAndServe(":" + socketPort, nil))
}

func Boot(pushPort string) *Sender {

	//  Socket to send messages on
	sender, err := zmq.NewSocket(zmq.PUSH)

	if err != nil {
		log.Fatal(err)
	}
	
	// Bind to push port
	sender.Bind("tcp://localhost:" + pushPort)

	spot := &Sender{sender}
	
	return spot
}