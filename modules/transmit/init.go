package transmit

import (
	"github.com/olebedev/config"
	"github.com/googollee/go-socket.io"
    "github.com/rs/cors"
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

        err = receiver.Bind("tcp://*:" + pullPort)

        if err != nil {
            panic(err)
        }

        log.Println("Started zmq pull server at tcp://*:" + pullPort)

        for {

            var message Message 
            
            msg, _ := receiver.Recv(0)

            if err := json.Unmarshal([]byte(msg), &message); err != nil {
                continue
            }

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
                log.Printf("%v\n", msg)

                server.BroadcastTo(msg.Room, msg.Event, msg.Message)
            }
        }()

        log.Println("Started sockets server at localhost:" + socketPort + "...")

        mux := http.NewServeMux()
        mux.Handle("/socket.io/", server)
        handler := cors.Default().Handler(mux)

        log.Fatal(http.ListenAndServe(":" + socketPort, handler))
    }()
	   
    log.Println("Waiting To Finish")
    wg.Wait()
}

func Boot(pushPort string) *Sender {

	spot := &Sender{pushPort}
	
	return spot
}