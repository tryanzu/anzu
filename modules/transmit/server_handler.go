package transmit

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/fernandez14/spartangeek-blacker/core/user"
	"github.com/googollee/go-socket.io"
	"gopkg.in/mgo.v2/bson"

	"encoding/json"
	"strings"
	"time"
)

// Handles a socket.io connection & performs some basic ops (auth, channel joining, etc.)
func handleConnection(deps Deps) func(so socketio.Socket) {
	log := deps.Log()
	redis := deps.Cache()
	secret, err := deps.Config().String("application.secret")

	if err != nil {
		log.Fatal("Could not get application secret. Can't handle auth.")
	}

	return func(so socketio.Socket) {
		token := so.Request().URL.Query().Get("token")

		if len(token) > 0 {
			signed, err := jwt.Parse(token, func(passed_token *jwt.Token) (interface{}, error) {
				// since we only use the one private key to sign the tokens,
				// we also only use its public counter part to verify
				return []byte(secret), nil
			})

			if err == nil {
				claims := signed.Claims.(jwt.MapClaims)
				sid := claims["user_id"].(string)
				oid := bson.ObjectIdHex(sid)
				usr, err := user.FindId(deps, oid)

				if err == nil {
					log.Debugf("Handled user %s connection.", usr.UserName)

					so.On("chat send", func(channel, message string) {
						message = strings.TrimSpace(message)

						if len(channel) < 1 || len(message) < 1 {
							return
						}

						if len(message) >= 200 {
							message = message[:200] + "..."
						}

						chat := map[string]interface{}{
							"list": []map[string]interface{}{
								{
									"content":   message,
									"user_id":   usr.Id,
									"username":  usr.UserName,
									"avatar":    usr.Image,
									"timestamp": time.Now().Unix(),
								},
							},
						}

						so.Emit("chat "+channel, chat)

						n, err := redis.Incr("chat:rates:m:" + usr.Id.Hex())
						if err != nil {
							log.Errorf("Error on rate limitter: %v", err)
							return
						}

						if n == 1 {
							redis.Expire("chat:rates:m:"+usr.Id.Hex(), 60)
						}

						if n > 10 {
							log.Debugf("Rate limit exceeded by %s", usr.UserName)
							return
						}

						if perSecond, err := redis.Exists("chat:rates:s:" + usr.Id.Hex()); perSecond && err == nil {
							log.Debugf("Rate limit exceeded by %s, no more than one message per second.", usr.UserName)
							return
						}

						so.BroadcastTo("chat", "chat "+channel, chat)

						_, err = redis.Incr("chat:rates:s:" + usr.Id.Hex())
						if err != nil {
							log.Error(err)
						}

						redis.Expire("chat:rates:s:"+usr.Id.Hex(), 1)

						// Async message saving.
						msg := Message{
							Room:    "chat",
							Event:   "chat " + channel,
							Message: chat,
						}

						if _, err := redis.LPush(msg.RoomID(), msg.Encode()); err != nil {
							log.Error("error:", err)
						}

						log.Debugf("Handling message %s to %s", message, channel)
					})
				}
			}
		} else {
			log.Debugf("Handled anonymous connection.")
		}

		so.Join("feed")
		so.Join("post")
		so.Join("general")
		so.Join("chat")
		so.Join("user")

		so.On("disconnection", func() {
			log.Debugf("Diconnection handled.")
		})

		so.On("chat update-me", func(channel string) {
			redis.LTrim("chat:"+channel, 0, 30)
			last, err := redis.LRange("chat:"+channel, 0, 30)
			if err == nil {
				// Allocate space for messages list.
				messages := Messages{
					List: []map[string]interface{}{},
				}

				for i := len(last) - 1; i >= 0; i-- {
					var m Message
					if err := json.Unmarshal([]byte(last[i]), &m); err != nil {
						continue
					}
					messages.List = append(messages.List, m.Message)
				}

				so.Emit("chat "+channel, messages)
			}
		})
	}
}
