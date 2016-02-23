package controller

import (
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"time"
)

type LeadAPI struct {
	Mongo  *mongo.Service               `inject:""`
}


func (this LeadAPI) Post(c *gin.Context) {

	var form LeadForm
	
	if c.BindJSON(&form) == nil {

		database := this.Mongo.Database
		similar, err := database.C("leads").Find(bson.M{"email": form.Email}).Count()

		if err != nil {
			panic(err)
		}

		var id bson.ObjectId

		if similar > 0 {

			var lead Lead

			err := database.C("leads").Find(bson.M{"email": form.Email}).One(&lead)

			if err != nil {
				panic(err)
			}

			err = database.C("leads").Update(bson.M{"email": form.Email}, bson.M{"$set": bson.M{"updated_at": time.Now()}, "$inc": bson.M{"seen": 1}})

			if err != nil {
				panic(err)
			}

			id = lead.Id

		} else {

			lead := Lead{
				Id: bson.NewObjectId(),
				Email: form.Email,
				Name: form.Name,
				Created: time.Now(),
			}

			err := database.C("leads").Insert(lead)

			if err != nil {
				panic(err)
			}

			id = lead.Id
		}


		c.JSON(200, gin.H{"status": "okay", "id": id.Hex()})
		return
	}

	c.JSON(400, gin.H{"status": "error", "message": "Invalid parameters."})
}

type Lead struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Email   string        `bson:"email" json:"email"`
	Name    string        `bson:"name" json:"name"`
	Created time.Time     `bson:"created_at" json:"created_at"`
}

type LeadForm struct {
	Email string `json:"email" binding:"required"`
	Name string `json:"name" binding:"required"`
}
