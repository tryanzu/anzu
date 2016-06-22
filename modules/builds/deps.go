package builds

import (
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"

	"time"
)

type Module struct {
	Mongo *mongo.Service `inject:""`
}

func (m *Module) FindOrCreate(sessionId string, userId *bson.ObjectId) *Build {

	var build Build

	db := m.Mongo.Database
	err := db.C("builds").Find(bson.M{"$or": []bson.M{{"session_id": sessionId}, {"user_id": userId}}}).One(&build)

	if err == nil {
		return &build
	}

	var ref string

	for {
		ref = helpers.StrRandom(6)
		count, err := db.C("builds").Find(bson.M{"ref": ref}).Count()

		if err != nil {
			panic(err)
		}

		if count > 0 {
			continue
		}

		break
	}

	build = Build{
		Id:        bson.NewObjectId(),
		Ref:       ref,
		SessionId: sessionId,
		Created:   time.Now(),
	}

	if userId != nil {
		build.UserId = userId
	}

	err = db.C("builds").Insert(build)

	if err != nil {
		panic(err)
	}

	return &build
}

func (m *Module) FindByRef(ref string) (*Build, error) {

	var build *Build

	err := m.Mongo.Database.C("builds").Find(bson.M{"ref": ref}).One(&build)

	if err != nil {
		return nil, err
	}

	build.SetDI(m)

	return build, nil
}
