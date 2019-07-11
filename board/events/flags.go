package events

import (
	"errors"
	"log"
	"time"

	"github.com/tryanzu/core/board/flags"
	"github.com/tryanzu/core/board/realtime"
	ev "github.com/tryanzu/core/core/events"
	"github.com/tryanzu/core/core/user"
	"github.com/tryanzu/core/deps"
	"gopkg.in/mgo.v2/bson"
)

// ErrInvalidIDRef for events with an id.
var ErrInvalidIDRef = errors.New("invalid id reference. could not find related object")

// Bind event handlers for flag related actions...
func flagHandlers() {
	ev.On <- ev.EventHandler{
		On: ev.NEW_FLAG,
		Handler: func(e ev.Event) error {
			fid := e.Params["id"].(bson.ObjectId)
			f, err := flags.FindId(deps.Container, fid)
			if err != nil {
				return ErrInvalidIDRef
			}
			if f.Reason == "spam" && f.RelatedTo == "chat" {
				usr, err := user.FindId(deps.Container, f.UserID)
				if err != nil {
					return err
				}
				ban, err := user.UpsertBan(deps.Container, user.Ban{
					UserID:    f.UserID,
					RelatedID: &fid,
					RelatedTo: "flag",
					Content:   "",
					Reason:    "spam",
				})
				if err != nil {
					return err
				}
				log.Println("[events] [flags] ban created with id", ban.ID)
				realtime.ToChan <- banLog("spam", "general", usr)
			}
			return nil
		},
	}
}

func banLog(reason, channel string, user user.User) realtime.M {
	return realtime.M{
		Channel: "chat:" + channel,
		Content: realtime.SocketEvent{
			Event: "log",
			Params: map[string]interface{}{
				"msg":  "%1$s has been banned for a while. reason: %2$s",
				"i18n": []string{user.UserName, reason},
				"meta": map[string]interface{}{
					"userId": user.Id,
					"user":   user.UserName,
					"reason": reason,
				},
				"at": time.Now(),
				"id": bson.NewObjectId(),
			},
		}.Encode(),
	}
}
