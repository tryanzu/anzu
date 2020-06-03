package realtime

import (
	"bytes"
	"encoding/gob"
	"html"
	"sync"
	"time"

	"github.com/desertbit/glue"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/tryanzu/core/board/flags"
	"github.com/tryanzu/core/core/content"
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

	seqHits  int
	lastRead *time.Time
}

func (c *Client) readWorker() {
	ledis := deps.Container.LedisDB()
	now := time.Now()
	c.seqHits = 0
	c.lastRead = &now
	for e := range c.Read {
		if c == nil || c.User != nil && user.IsBanned(deps.Container, c.User.Id) {
			continue
		}
		switch e.Event {
		case "auth":
			c.readAuth(e)
		case "auth:clean":
			c.readAuthClean(e)
		case "listen":
			c.readListen(e)
		case "unlisten":
			channel, exists := e.Params["chan"].(string)
			if !exists {
				log.Debugf("Could not remove channel: missing id")
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
		case "chat:message":
			c.readChatMessage(e)
		case "chat:delete":
			if c.User == nil {
				continue
			}
			mid, exists := e.Params["id"].(string)
			if !exists || bson.IsObjectIdHex(mid) == false {
				log.Warning("chat:delete requires a valid message id.")
				continue
			}
			channel, exists := e.Params["chan"].(string)
			if !exists {
				log.Debugf("chat:message requires a chan.")
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
				log.Debugf("chat:ban requires a higher privileges.")
				continue
			}
			uid, exists := e.Params["userId"].(string)
			if !exists || bson.IsObjectIdHex(uid) == false {
				log.Debugf("chat:ban requires a valid user id.")
				continue
			}
			events.In <- events.NewBanFlag(bson.ObjectIdHex(uid))
		case "chat:star":
			if c.User == nil {
				continue
			}
			msg, exists := e.Params["message"].(map[string]interface{})
			if !exists {
				log.Debugf("chat:star requires a valid message.")
				continue
			}
			channel, exists := e.Params["chan"].(string)
			if !exists {
				log.Debugf("chat:message requires a chan.")
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
		}
	}
	log.Infof("read worker stopped, client = %+v", c)
}

// finish client connection.
func (c *Client) finish() {
	var uid *bson.ObjectId
	if c.User != nil {
		uid = &c.User.Id
		err := user.LastSeenAt(deps.Container, c.User.Id, time.Now())
		if err != nil {
			log.Error(err)
		}
	}
	// Close the channel so readWorker stops.
	close(c.Read)
	addresses.Delete(c.Raw.RemoteAddr())
	log.Infof("socket closed	id = %s | address = %s | userId = %v", c.Raw.ID(), c.Raw.RemoteAddr(), uid)

	// Clean up pointers & logging.
	c.User = nil
	c.Channels = nil
	sockets.Delete(c.Raw.ID())
	c.Raw = nil

	// Acknowledge connected client in counters.
	counters <- c
}

func (c *Client) readAuth(e SocketEvent) {
	token, exists := e.Params["token"].(string)
	if !exists {
		log.Warning("could not authenticate socket client: missing token")
		return
	}

	signed, err := jwt.Parse(token, func(passed_token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		log.Warningf("could not parse socket client token: %v", err)
		return
	}

	claims := signed.Claims.(jwt.MapClaims)
	usr, err := user.FindId(deps.Container, bson.ObjectIdHex(claims["user_id"].(string)))
	if err != nil {
		log.Errorf("could not find user from socket token: %v", err)
		return
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
}

func (c *Client) readAuthClean(e SocketEvent) {
	c.User = nil
	c.SafeWrite(SocketEvent{
		Event: "auth:cleaned",
	}.encode())
	counters <- c
}

type chatMessage map[string]interface{}

func (c chatMessage) GetContent() string {
	return c["msg"].(string)
}

func (c chatMessage) UpdateContent(content string) content.Parseable {
	c["msg"] = content
	return c
}

func (c chatMessage) GetParseableMeta() map[string]interface{} {
	meta := make(map[string]interface{})
	meta["id"] = c["id"]
	meta["related"] = "chat"
	meta["user_id"] = c["userId"]
	return meta
}

func (c *Client) readChatMessage(e SocketEvent) {
	if c.User == nil {
		return
	}
	msg, exists := e.Params["msg"].(string)
	if !exists || len(msg) == 0 || len(msg) > 255 {
		log.Debugf("chat:message requires a valid message.")
		return
	}
	var channel string
	channel, exists = e.Params["chan"].(string)
	if !exists {
		log.Warning("chat:message requires a chan.")
		return
	}
	mid := bson.NewObjectId()
	chatM := chatMessage{
		"msg":    html.EscapeString(msg),
		"userId": c.User.Id,
		"from":   c.User.UserName,
		"avatar": c.User.Image,
		"at":     time.Now(),
		"id":     mid,
	}
	pre, err := content.Preprocess(deps.Container, chatM)
	if err != nil {
		log.Errorf("could not preprocess chat message, err: %v", err)
		return
	}
	chatM = pre.(chatMessage)
	post, err := content.Postprocess(deps.Container, chatM)
	if err != nil {
		log.Errorf("could not postprocess chat message, err: %v", err)
		return
	}
	chatM = post.(chatMessage)
	m := M{
		ID:      &mid,
		Channel: "chat:" + channel,
		Content: SocketEvent{
			Event:  "message",
			Params: chatM,
		}.encode(),
	}
	ToChan <- m
	now := time.Now()
	last := *c.lastRead
	log.Debugf("client chat message, lastRead = %v", now.Sub(last))
	if now.Sub(last) < 300*time.Millisecond {
		c.seqHits++
	} else {
		c.seqHits = 0
	}

	// Update last input time from client.
	c.lastRead = &now
	if c.seqHits >= 10 {
		log.Infof("spam rate exceeded, sending sys flag, client = %+v", c)
		c.sysFlag("spam")
	}
	time.Sleep(time.Millisecond * 60)
	mark := elapsed("caching message")
	ledisdb := deps.Container.LedisDB()
	var (
		buf bytes.Buffer
		n   int64
	)
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(m)
	if err != nil {
		log.Debug("[err] Cannot encode for cache", err)
	}
	n, err = ledisdb.RPush([]byte(m.Channel), buf.Bytes())
	if err != nil {
		log.Debug("[err] Cannot encode for cache", err)
	}
	if n >= 50 {
		ledisdb.LPop([]byte(m.Channel))
	}
	mark()
}

func (c *Client) readListen(e SocketEvent) {
	channel, exists := e.Params["chan"].(string)
	if !exists {
		log.Warning("could not join channel: missing id")
		return
	}
	c.Channels.Store(channel, c.Raw.Channel(channel))
	c.SafeWrite(SocketEvent{
		Event: "listen:ready",
		Params: map[string]interface{}{
			"chan": channel,
		},
	}.encode())

	// When a user listens to a chat channel additional business logic needs to be executed
	if channel[0:4] == "chat" {
		err := c.enterChatChannel(channel)
		if err != nil {
			// switch err.(type) {
			// case :

			// }
		}
	}

	// Acknowledge connected client in counters.
	counters <- c
}

func (c *Client) enterChatChannel(channel string) error {
	ledis := deps.Container.LedisDB()
	prev, err := ledis.LRange([]byte(channel), 0, 50)
	if err != nil {
		log.Error(err)
		return nil
	}
	for _, encoded := range prev {
		var msg M
		dec := gob.NewDecoder(bytes.NewBuffer(encoded))
		err := dec.Decode(&msg)
		if err != nil {
			log.Warningf("cannot decode previous chat message, err is: %v", err)
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
	return nil
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
		log.Errorf("sysFlag failed, userId = %s | relatedTo: chat | reason: %s", c.User.Id, reason)
		log.Error(err)
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
				log.Debugf("invalid userId in packed messages sending, chan = %s", m.Channel)
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
