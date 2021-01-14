package events

import (
	"errors"
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
					RelatedTo: "chat",
					Content:   "Flag received from chat",
					Reason:    "spam",
				})
				if err != nil {
					return err
				}
				log.Debug("ban created with id", ban.ID)
				realtime.ToChan <- banLog(ban, usr)
			}
			return nil
		},
	}
	ev.On <- ev.EventHandler{
		On: ev.NEW_BAN,
		Handler: func(e ev.Event) error {
			uid := e.Params["userId"].(bson.ObjectId)
			usr, err := user.FindId(deps.Container, uid)
			if err != nil {
				return err
			}
			ban, err := user.UpsertBan(deps.Container, user.Ban{
				UserID:    uid,
				RelatedID: &uid,
				RelatedTo: "chat",
				Content:   "Flag ban received from chat",
				Reason:    "spam",
			})
			if err != nil {
				return err
			}
			log.Debug("ban created with id", ban.ID)
			realtime.ToChan <- banLog(ban, usr)
			return nil
		},
	}
}

func banLog(ban user.Ban, user user.User) realtime.M {
	diff := ban.Until.Sub(ban.Created).Truncate(time.Second)
	return realtime.M{
		Channel: "chat:general",
		Content: realtime.SocketEvent{
			Event: "log",
			Params: map[string]interface{}{
				"msg":  "%1$s has been banned for %3$s. reason: %2$s",
				"i18n": []string{user.UserName, ban.Reason, diff.String()},
				"meta": map[string]interface{}{
					"userId": user.Id,
					"user":   user.UserName,
					"reason": ban.Reason,
				},
				"at": ban.Created,
				"id": bson.NewObjectId(),
			},
		}.Encode(),
	}
}
