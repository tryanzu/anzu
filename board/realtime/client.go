package realtime

import (
	"log"

	"github.com/desertbit/glue"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

// Client contains a message to be broadcasted to a channel
type Client struct {
	Raw      *glue.Socket
	Channels map[string]*glue.Channel
	User     *user.User
	Read     chan socketEvent
}

func (c *Client) readWorker() {
	for e := range c.Read {
		switch e.Event {
		case "auth":
			token, exists := e.Params["token"].(string)
			if !exists {
				log.Println("Could not authenticate socket client: missing token")
				continue
			}

			signed, err := jwt.Parse(token, func(passed_token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})

			if err != nil {
				log.Println("Could not parse socket client token: ", err)
				continue
			}

			claims := signed.Claims.(jwt.MapClaims)
			usr, err := user.FindId(deps.Container, bson.ObjectIdHex(claims["user_id"].(string)))
			if err != nil {
				log.Println("Could not find user from socket token: ", err)
				continue
			}

			c.User = &usr
			event := socketEvent{
				Event: "auth:my",
				Params: map[string]interface{}{
					"user": usr,
				},
			}

			c.Raw.Write(event.encode())
		case "auth:clean":
			c.User = nil
			c.Raw.Write(socketEvent{
				Event: "auth:cleaned",
			}.encode())
		}
	}
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
