package transmit

import (
	"github.com/olebedev/config"
	"github.com/googollee/go-socket.io"
	zmq "github.com/pebbe/zmq4"

    "log"
    "net/http"
    "encoding/json"
    "sync"
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

    var wg sync.WaitGroup
    wg.Add(2)

    messages := make(chan Message)

    go func() {

        defer wg.Done()

        //  Socket to receive messages on
        receiver, err := zmq.NewSocket(zmq.PULL)

        if err != nil {
            panic(err)
        }

        defer receiver.Close()
        receiver.Bind("tcp://127.0.0.1:" + pullPort)

        for {

            var message Message 
            
            msg, _ := receiver.Recv(0)

            if err := json.Unmarshal([]byte(msg), &message); err != nil {
                continue
            }

            messages <- message

            log.Println("Broadcasted message to " + message.Room)
        }
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
            so.Join("user")
            
            so.On("disconnection", func() {
                log.Println("Diconnection handled.")
            })
        })

        server.On("error", func(so socketio.Socket, err error) {
            log.Println("error:", err)
        })

        go func() {
            for {
                msg := <-messages
                log.Println("received message from channel")
                log.Printf("%v\n", msg)

                server.BroadcastTo(msg.Room, msg.Event, msg.Message)
            }
        }()

        log.Println("Started sockets server at localhost:" + socketPort + "...")
        http.Handle("/socket.io/", server)
        log.Fatal(http.ListenAndServe(":" + socketPort, nil))
    }()
	   
    log.Println("Waiting To Finish")
    wg.Wait()
}

func Boot(pushPort string) *Sender {

	spot := &Sender{pushPort}
	
	return spot
}