package post

import (
	"context"
	"errors"
	"fmt"
	"github.com/matcornic/hermes/v2"
	"github.com/tryanzu/core/core/config"
	"math"

	"github.com/tryanzu/core/core/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PostNotFound err.
var PostNotFound = errors.New("post has not been found by given criteria")

func FindId(deps deps, id primitive.ObjectID) (post Post, err error) {
	ctx := context.TODO()
	err = deps.Mgo().Collection("posts").FindOne(ctx, bson.M{"_id": id}).Decode(&post)
	if err == mongo.ErrNoDocuments {
		err = PostNotFound
	}
	return
}

func FindList(deps deps, scopes ...common.Scope) (list Posts, err error) {
	ctx := context.TODO()
	cursor, err := deps.Mgo().Collection("posts").Find(ctx, common.ByScope(scopes...))
	if err != nil {
		return
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &list)
	return
}

func FindRateList(d deps, date string, offset, limit int) ([]primitive.ObjectID, error) {
	list := []primitive.ObjectID{}
	scores, err := d.LedisDB().ZRangeByScoreGeneric([]byte("posts:"+date), 0, math.MaxInt64, offset, limit, true)
	if err != nil {
		return list, err
	}
	for _, n := range scores {
		id, err := primitive.ObjectIDFromHex(string(n.Member))
		if err != nil {
			continue
		}
		list = append(list, id)
	}
	log.Info("getting rate list at %s", date)
	return list, err
}

func SomeoneCommentedYourPost(name string, post Post) hermes.Email {
	c := config.C.Copy()
	link := c.Site.MakeURL("p/" + post.Slug + "/" + post.Id.Hex())
	return hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				fmt.Sprintf("Tu publicación en %s (%s) recibió un comentario mientras no estabas.", c.Site.Name, post.Title),
			},
			Actions: []hermes.Action{
				{
					Button: hermes.Button{
						Color: "#3D5AFE",
						Text:  "Ver publicación",
						Link:  link,
					},
				},
			},
			Outros: []string{
				"Si deseas dejar de recibir notificaciones puedes entrar en tu cuenta y cambiar la configuración de avisos.",
			},
			Signature: "Un saludo",
		},
	}
}
