package payments

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"strconv"
	"time"
)

func (this API) PaypalExecute(c *gin.Context) {

	var m ExecutePayload

	if c.Bind(&m) == nil {

		var saved Payment
		database := this.Mongo.Database
		err := database.C("payments").Find(bson.M{"gateway": "paypal", "gateway_id": m.PaymentId}).One(&saved)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "not-found", "details": err})
		}

		if saved.Status != "created" {
			c.JSON(400, gin.H{"status": "error", "error": "invalid-action", "details": "Can't execute this payment."})
			return
		}

		client := this.GetPaypalClient()
		payment, err := client.ExecutePayment(m.PaymentId, m.PayerId, nil)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "execution-failed", "details": err})
			return
		}

		err = database.C("payments").Update(bson.M{"_id": saved.Id}, bson.M{"$set": bson.M{"updated_at": time.Now(), "status": "confirmed"}})

		if err != nil {
			panic(err)
		}

		id := c.MustGet("user_id")
		user_id := bson.ObjectIdHex(id.(string))
		usr := this.Gaming.Get(user_id)

		if err != nil {
			panic(err)
		}

		var total float64

		total, err = strconv.ParseFloat(payment.Transactions[0].Amount.Total, 64)

		if err != nil {
			panic(err)
		}

		// Gift user a badge - TODO: find better way for this matter
		if total >= 50 && total < 200 {
			usr.AcquireBadge(bson.ObjectIdHex("55f0a94fbe6bceb9b3762c6c"), false)
		} else if total >= 200 && total < 1000 {
			usr.AcquireBadge(bson.ObjectIdHex("55f0a950be6bceb9b3762c6d"), false)
		} else if total > 1000 {
			usr.AcquireBadge(bson.ObjectIdHex("55f0a951be6bceb9b3762c6e"), false)
		}

		c.JSON(200, gin.H{"status": "okay", "response": payment})
	}
}
