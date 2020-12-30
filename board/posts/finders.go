package post

import (
	"errors"
	"fmt"
	"github.com/matcornic/hermes/v2"
	"github.com/tryanzu/core/core/config"
	"math"

	"github.com/tryanzu/core/core/common"
	"gopkg.in/mgo.v2/bson"
)

// PostNotFound err.
var PostNotFound = errors.New("post has not been found by given criteria")

func FindId(deps deps, id bson.ObjectId) (post Post, err error) {
	err = deps.Mgo().C("posts").FindId(id).One(&post)
	return
}

func FindList(deps deps, scopes ...common.Scope) (list Posts, err error) {
	err = deps.Mgo().C("posts").Find(common.ByScope(scopes...)).All(&list)
	return
}

func FindRateList(d deps, date string, offset, limit int) ([]bson.ObjectId, error) {
	list := []bson.ObjectId{}
	scores, err := d.LedisDB().ZRangeByScoreGeneric([]byte("posts:"+date), 0, math.MaxInt64, offset, limit, true)
	if err != nil {
		return list, err
	}
	for _, n := range scores {
		id := bson.ObjectIdHex(string(n.Member))
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
