package realtime

import (
	"bytes"
	"encoding/gob"
	"log"
	"sync"
	"time"

	"github.com/desertbit/glue"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/tryanzu/core/board/flags"
	"github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

// Client contains a message to be broadcasted to a channel
type Client struct {
	Raw      *glue.Socket
	Channels *sync.Map
	// Channels map[string]*glue.Channel
	User *user.User
	Read chan SocketEvent
}

func (c *Client) readWorker() {
	ledis := deps.Container.LedisDB()
	seqHits := 0
	lastRead := time.Now()
	for e := range c.Read {
		if c == nil || c.User != nil && user.IsBanned(deps.Container, c.User.Id) {
			continue
		}
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
			event := SocketEvent{
				Event: "auth:my",
				Params: map[string]interface{}{
					"user": usr,
				},
			}
			c.SafeWrite(event.encode())
			counters <- c

		case "auth:clean":
			c.User = nil
			c.SafeWrite(SocketEvent{
				Event: "auth:cleaned",
			}.encode())
			counters <- c

		case "listen":
			channel, exists := e.Params["chan"].(string)
			if !exists {
				log.Println("Could not join channel: missing id")
				continue
			}

			c.Channels.Store(channel, c.Raw.Channel(channel))
			c.SafeWrite(SocketEvent{
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
					if msg.ID != nil {
						n, err := ledis.SIsMember([]byte(channel+":deleted"), []byte(*msg.ID))
						if n == 1 || err != nil {
							continue
						}
					}
					if ch, ok := c.Channels.Load(channel); ok && c != nil {
						ch.(*glue.Channel).Write(msg.Content)
					}
				}
			}
			counters <- c

		case "unlisten":
			channel, exists := e.Params["chan"].(string)
			if !exists {
				log.Println("Could not remove channel: missing id")
				continue
			}
			c.Channels.Delete(channel)
			c.SafeWrite(SocketEvent{
				Event: "unlisten:ready",
				Params: map[string]interface{}{
					"chan": channel,
				},
			}.encode())
			counters <- c
		case "chat:delete":
			if c.User == nil {
				continue
			}
			mid, exists := e.Params["id"].(string)
			if !exists || bson.IsObjectIdHex(mid) == false {
				log.Println("[glue] chat:delete requires a valid message id.")
				continue
			}
			channel, exists := e.Params["chan"].(string)
			if !exists {
				log.Println("[glue] chat:message requires a chan.")
				continue
			}
			m := M{
				Channel: "chat:" + channel,
				Content: SocketEvent{
					Event: "delete",
					Params: map[string]interface{}{
						"id": mid,
					},
				}.encode(),
			}
			ledis.SAdd([]byte(m.Channel+":deleted"), []byte(bson.ObjectIdHex(mid)))
			ToChan <- m
		case "chat:ban":
			if c.User == nil {
				continue
			}
			if c.User.HasRole("admin", "developer") == false {
				log.Println("[glue] chat:ban requires a higher privileges.")
				continue
			}
			uid, exists := e.Params["userId"].(string)
			if !exists || bson.IsObjectIdHex(uid) == false {
				log.Println("[glue] chat:ban requires a valid user id.")
				continue
			}
			events.In <- events.NewBanFlag(bson.ObjectIdHex(uid))
		case "chat:star":
			if c.User == nil {
				continue
			}
			msg, exists := e.Params["message"].(map[string]interface{})
			if !exists {
				log.Println("[glue] chat:star requires a valid message.")
				continue
			}
			channel, exists := e.Params["chan"].(string)
			if !exists {
				log.Println("[glue] chat:message requires a chan.")
				continue
			}
			go func() {
				m := M{
					Channel: "chat:" + channel,
					Content: SocketEvent{
						Event:  "star",
						Params: msg,
					}.encode(),
				}
				featuredM <- m
			}()
		case "chat:message":
			if c.User == nil {
				continue
			}
			msg, exists := e.Params["msg"].(string)
			if !exists || len(msg) == 0 || len(msg) > 255 {
				log.Println("[glue] chat:message requires a valid message.")
				continue
			}
			var channel string
			channel, exists = e.Params["chan"].(string)
			if !exists {
				log.Println("[glue] chat:message requires a chan.")
				continue
			}
			mid := bson.NewObjectId()
			m := M{
				ID:      &mid,
				Channel: "chat:" + channel,
				Content: SocketEvent{
					Event: "message",
					Params: map[string]interface{}{
						"msg":    msg,
						"userId": c.User.Id,
						"from":   c.User.UserName,
						"avatar": c.User.Image,
						"at":     time.Now(),
						"id":     mid,
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

			// Update last input time from client.
			lastRead = time.Now()
			if seqHits >= 10 {
				log.Println("[glue] [ban] spam rate exceeded, sending flag")
				c.sysFlag("spam")
			}
			time.Sleep(time.Millisecond * 200)
		}
	}
}

func (c *Client) sysFlag(reason string) {
	if c.User == nil {
		return
	}
	flag, err := flags.UpsertFlag(deps.Container, flags.Flag{
		UserID:    c.User.Id,
		RelatedTo: "chat",
		Content:   "System has sent this flag.",
		Reason:    reason,
	})
	if err != nil {
		log.Println("[glue] [error] sysFlag failed, error:", err)
		return
	}
	events.In <- events.NewFlag(flag.ID)
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

		if c, exists := c.Channels.Load(m.Channel); exists && c != nil {
			c.(*glue.Channel).Write(m.Content)
		}
	}
}
