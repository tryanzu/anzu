package main

import (
    "fmt"
    "encoding/json"
    zmq "github.com/pebbe/zmq4"
)

func main() {
    
    socket, err  := zmq.NewSocket(zmq.REQ)
    
    if err != nil { panic(err) }
    
    socket.Connect("tcp://127.0.0.1:5439")
    
    for i := 0; i < 20; i++ {
       
        msg := fmt.Sprintf("msg %d", i)
        
        testMessage := map[string] string {
            "title": "A test title",
            "message": msg,
            "to": "messaging:main",
        }

        test, _ := json.Marshal(testMessage)
        
        socket.Send(string(test), 0)
        println("Sending", string(test))
    	
        got, _ := socket.Recv(0)
        
        println("got ", string(got))
  }
}