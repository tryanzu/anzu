package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/gin-gonic/gin"
	messagebird "github.com/messagebird/go-rest-api"
	"gopkg.in/mgo.v2/bson"

	"fmt"
	"time"
)

type LeadAPI struct {
	Mongo *mongo.Service `inject:""`
	User  *user.Module   `inject:""`
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
				Id:      bson.NewObjectId(),
				Email:   form.Email,
				Name:    form.Name,
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

func (container LeadAPI) GetContestLead(c *gin.Context) {
	var original ContestLead

	db := container.Mongo.Database
	user_str := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(user_str.(string))

	// Find record if any for further diffs
	db.C("contest_leads").Find(bson.M{"user_id": user_id}).One(&original)

	original.Code = "secret"

	c.JSON(200, original)
}

func (container LeadAPI) UpdateContestLead(c *gin.Context) {
	var form, original ContestLead

	db := container.Mongo.Database
	user_str := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(user_str.(string))
	usr, err := container.User.Get(user_id)
	if err != nil {
		panic(err)
	}

	// Find record if any for further diffs
	db.C("contest_leads").Find(bson.M{"user_id": user_id}).One(&original)

	if c.BindJSON(&form) == nil {
		// Upsert that record with new data
		var code string

		if len(original.Code) > 0 {
			code = original.Code
		} else {
			code = helpers.StrNumRandom(5)
		}

		isNew := !original.Id.Valid()
		_, err = container.Mongo.Database.C("contest_leads").Upsert(bson.M{"user_id": user_id}, bson.M{
			"$set": bson.M{
				"step":       form.Step,
				"email":      form.Email,
				"name":       form.Name,
				"phone":      form.Phone,
				"birthday":   form.Birthday,
				"additional": form.Additional,
			},
			"$setOnInsert": bson.M{
				"code":       code,
				"created_at": time.Now(),
				"validated":  false,
			},
		})

		if err != nil {
			panic(err)
		}

		sent := false
		_, resend := c.GetQuery("resend")

		if original.SentTimes < 3 && (original.Phone != form.Phone || resend) && len(form.Phone) == 10 {
			confirm := fmt.Sprintf("Hola %s! Tu codigo de confirmación es %s", usr.Name(), code)

			client := messagebird.New("OVa8Syl3B47DfdRgiahMvd6KV")
			message, err := client.NewMessage("spartangeek", []string{"+521" + form.Phone}, confirm, nil)

			if err != nil {
				c.JSON(422, gin.H{"status": "error", "message": "No pudimos enviarte un código de confirmación a ese número.", "details": err})
				return
			}

			err = db.C("contest_leads").Update(bson.M{"user_id": user_id}, bson.M{"$set": bson.M{"last_sent": message.CreatedDatetime}, "$inc": bson.M{"sent_times": 1}})
			if err != nil {
				panic(err)
			}

			sent = true
		}

		if !isNew && code == form.Code && original.Validated == false {
			err = db.C("contest_leads").Update(bson.M{"user_id": user_id}, bson.M{"$set": bson.M{"validated": true, "validated_at": time.Now()}})
			if err != nil {
				panic(err)
			}
		}

		c.JSON(200, gin.H{"status": "okay", "sms_sent": sent})
	}
}

type ContestLead struct {
	Id          bson.ObjectId          `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectId          `bson:"user_id" json:"user_id"`
	Step        int                    `bson:"step" json:"step"`
	Email       string                 `bson:"email" json:"email"`
	Name        string                 `bson:"name" json:"name"`
	Phone       string                 `bson:"phone" json:"phone"`
	Birthday    string                 `bson:"birthday" json:"birthday"`
	Validated   bool                   `bson:"validated" json:"validated"`
	ValidatedAt time.Time              `bson:"validated_at" json:"validated_at"`
	Code        string                 `bson:"code" json:"code"`
	SentTimes   int                    `bson:"sent_times" json:"sent_times"`
	LastSent    time.Time              `bson:"last_sent" json:"last_sent"`
	Additional  map[string]interface{} `bson:"additional" json:"additional"`
	Created     time.Time              `bson:"created_at" json:"created_at"`
}

type Lead struct {
	Id      bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Email   string        `bson:"email" json:"email"`
	Name    string        `bson:"name" json:"name"`
	Created time.Time     `bson:"created_at" json:"created_at"`
}

type LeadForm struct {
	Email string `json:"email" binding:"required"`
	Name  string `json:"name" binding:"required"`
}
