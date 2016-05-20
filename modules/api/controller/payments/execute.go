package payments

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

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

		c.JSON(200, gin.H{"status": "okay", "response": payment})
	}
}
