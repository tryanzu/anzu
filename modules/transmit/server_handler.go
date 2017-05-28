package transmit

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/fernandez14/spartangeek-blacker/core/user"
	"github.com/googollee/go-socket.io"
	"gopkg.in/mgo.v2/bson"

	"strings"
	"time"
)

type MessageBuilder func(string) map[string]interface{}

type ChannelMessage struct {
	Channel string
	Message string
}

type PackedMessage struct {
	Channel string
	Message map[string]interface{}
}

// Anonymous message builder.
func anonymousMessage(str string) map[string]interface{} {
	str = strings.TrimSpace(str)

	if len(str) >= 200 {
		str = str[:200] + "..."
	}

	return map[string]interface{}{
		"content":        str,
		"user_id":        "guest",
		"username":       "guest",
		"avatar":         false,
		"timestamp":      time.Now().Unix(),
		"timestamp_nano": time.Now().UnixNano(),
	}
}

// Handle socket request authentication token.
func handleTokenAuth(token, secret string, deps Deps) MessageBuilder {
	if len(token) == 0 {
		return anonymousMessage
	}

	signed, err := jwt.Parse(token, func(passed_token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return []byte(secret), nil
	})

	if err != nil {
		return anonymousMessage
	}

	claims := signed.Claims.(jwt.MapClaims)
	sid := claims["user_id"].(string)
	oid := bson.ObjectIdHex(sid)
	usr, err := user.FindId(deps, oid)
	if err != nil {
		return anonymousMessage
	}

	return func(message string) map[string]interface{} {
		message = strings.TrimSpace(message)

		if len(message) >= 200 {
			message = message[:200] + "..."
		}

		return map[string]interface{}{
			"content":        message,
			"user_id":        usr.Id,
			"username":       usr.UserName,
			"avatar":         usr.Image,
			"timestamp":      time.Now().Unix(),
			"timestamp_nano": time.Now().UnixNano(),
		}
	}
}

// Pack a list of messages.
func list(messages ...map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"list": messages,
	}
}

// Handles a socket.io connection & performs some basic ops (auth, channel joining, etc.)
func handleConnection(deps Deps) func(so socketio.Socket) {
	log := deps.Log()
	mgo := deps.Mgo()
	secret, err := deps.Config().String("application.secret")

	if err != nil {
		log.Fatal("Could not get application secret. Can't handle auth.")
	}

	historic := map[string][]map[string]interface{}{}
	history := make(chan PackedMessage, 100)
	registry := make(chan PackedMessage, 100)

	// History buffer consumer.
	go func() {
		for {
			h := <-history

			if _, exists := historic[h.Channel]; !exists {
				historic[h.Channel] = []map[string]interface{}{}
			}

			// Shift first item in the historic
			if len(historic[h.Channel]) >= 30 {
				historic[h.Channel] = historic[h.Channel][1:]
			}

			historic[h.Channel] = append(historic[h.Channel], h.Message)

			go func() {
				select {
				case registry <- h:
					log.Debug("Chat message got to the registry.")
				case <-time.After(time.Second * 5):
					log.Critical("Could not get chat to the registry.")
				}
			}()
		}
	}()

	go func() {
		for {
			r := <-registry
			ts := r.Message["timestamp_nano"].(int64)
			message := r.Message
			message["lag"] = time.Now().UnixNano() - ts
			err := mgo.C("chat_messages").Insert(message)

			if err != nil {
				log.Criticalf("Could not log chat message: %v", err)
			}
		}
	}()

	return func(so socketio.Socket) {
		token := so.Request().URL.Query().Get("token")
		builder := handleTokenAuth(token, secret, deps)

		// Messaging buffer holding pending messages to broadcast progressively
		messaging := make(chan ChannelMessage, 10)

		// Messaging buffer consumer.
		go func() {
			for {
				m := <-messaging

				// Build message to be sent over the wire.
				message := builder(m.Message)
				packed := list(message)

				so.Emit("chat "+m.Channel, packed)
				so.BroadcastTo("chat", "chat "+m.Channel, packed)

				// Send to history
				history <- PackedMessage{m.Channel, message}

				// Rate limit.
				time.Sleep(time.Second)
			}
		}()

		so.On("chat send", func(channel, message string) {
			m := ChannelMessage{
				Channel: channel,
				Message: message,
			}

			messaging <- m
		})

		so.On("chat update-me", func(channel string) {
			if hlist, exists := historic[channel]; exists {
				packed := list(hlist...)
				so.Emit("chat "+channel, packed)
			}
		})

		so.Join("feed")
		so.Join("post")
		so.Join("general")
		so.Join("chat")
		so.Join("user")

		so.On("disconnection", func() {
			log.Debugf("Diconnection handled.")
		})
	}
}
