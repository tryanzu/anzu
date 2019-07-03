package realtime

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

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
	ledis := deps.Container.LedisDB()
	seqHits := 0
	lastRead := time.Now()
	for e := range c.Read {
		switch e.Event {
		case "auth":
			token, exists := e.Params["token"].(string)
			if !exists {
				log.Println("[REALTIME] Could not authenticate socket client: missing token")
				continue
			}

			signed, err := jwt.Parse(token, func(passed_token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})

			if err != nil {
				log.Println("[REALTIME] Could not parse socket client token: ", err)
				continue
			}

			claims := signed.Claims.(jwt.MapClaims)
			usr, err := user.FindId(deps.Container, bson.ObjectIdHex(claims["user_id"].(string)))
			if err != nil {
				log.Println("[REALTIME] Could not find user from socket token: ", err)
				continue
			}

			c.User = &usr
			event := socketEvent{
				Event: "auth:my",
				Params: map[string]interface{}{
					"user": usr,
				},
			}
			c.SafeWrite(event.encode())

		case "auth:clean":
			c.User = nil
			c.SafeWrite(socketEvent{
				Event: "auth:cleaned",
			}.encode())

		case "listen":
			channel, exists := e.Params["chan"].(string)
			if !exists {
				log.Println("Could not join channel: missing id")
				continue
			}

			c.Channels[channel] = c.Raw.Channel(channel)
			c.SafeWrite(socketEvent{
				Event: "listen:ready",
				Params: map[string]interface{}{
					"chan": channel,
				},
			}.encode())
			if channel[0:4] == "chat" {
				prev, err := ledis.LRange([]byte(channel), 0, 50)
				if err != nil {
					log.Println("[glue] [err] Cannot get previous chat list", err)
					continue
				}
				for _, encoded := range prev {
					var msg M
					dec := gob.NewDecoder(bytes.NewBuffer(encoded))
					err := dec.Decode(&msg)
					if err != nil {
						log.Println("[glue] [err] Cannot decode previous chat message", err)
						continue
					}
					c.Channels[channel].Write(msg.Content)
				}
			}
			counters <- c

		case "unlisten":
			channel, exists := e.Params["chan"].(string)
			if !exists {
				log.Println("Could not remove channel: missing id")
				continue
			}
			delete(c.Channels, channel)
			c.SafeWrite(socketEvent{
				Event: "unlisten:ready",
				Params: map[string]interface{}{
					"chan": channel,
				},
			}.encode())
			counters <- c
		case "chat:message":
			if c.User == nil {
				continue
			}
			msg, exists := e.Params["msg"].(string)
			if !exists || len(msg) == 0 {
				log.Println("[glue] chat:message requires a message.")
				continue
			}
			var channel string
			channel, exists = e.Params["chan"].(string)
			if !exists {
				log.Println("[glue] chat:message requires a chan.")
				continue
			}
			m := M{
				Channel: "chat:" + channel,
				Content: socketEvent{
					Event: "message",
					Params: map[string]interface{}{
						"msg":    msg,
						"userId": c.User.Id,
						"from":   c.User.UserName,
						"avatar": c.User.Image,
						"at":     time.Now(),
						"id":     bson.NewObjectId(),
					},
				}.encode(),
			}
			ToChan <- m
			t := time.Now()
			log.Println("[glue] lastRead ", t.Sub(lastRead))
			if t.Sub(lastRead) < 300*time.Millisecond {
				seqHits++
			} else {
				seqHits = 0
			}
			lastRead = time.Now()
			var (
				network bytes.Buffer
				n       int64
			)
			enc := gob.NewEncoder(&network)
			err := enc.Encode(m)
			if err != nil {
				log.Println("[glue] [err] Cannot encode for cache", err)
			}
			n, err = ledis.RPush([]byte(m.Channel), network.Bytes())
			if err != nil {
				log.Println("[glue] [err] Cannot encode for cache", err)
			}
			if n >= 50 {
				ledis.LPop([]byte(m.Channel))
			}
			if seqHits >= 10 {
				log.Println("[glue] [ban] spam rate exceeded")
				time.Sleep(time.Second * 60)
			}
			time.Sleep(time.Millisecond * 200)
		}
	}
}

// SafeWrite to client (from nil pointers)
func (c *Client) SafeWrite(data string) {
	if c.Raw != nil {
		c.Raw.Write(data)
	}
}

func (c *Client) send(packed []M) {
	for _, m := range packed {
		if m.Channel == "" {
			c.SafeWrite(m.Content)
			continue
		}
		if m.Channel[0:4] == "user" {
			if c.User == nil {
				continue
			}
			id := bson.ObjectIdHex(m.Channel[5:])
			if id.Valid() == false {
				log.Println("Invalid userId in packed messages sending. Chan:", m.Channel)
				continue
			}
			if id == c.User.Id {
				c.SafeWrite(m.Content)
			}
		}

		if c, exists := c.Channels[m.Channel]; exists {
			c.Write(m.Content)
		}
	}
}
