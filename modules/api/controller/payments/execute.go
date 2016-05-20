package payments

import (
	"github.com/gin-gonic/gin"
	"github.com/leebenson/paypal"
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

		clientID := "ASkcCWSuhGXVNim-dfAJ9Zlxk41iLceeLLSv7_dvBZY-Dob1sGBVFaMgIUKaOyHb9TmjWXgV83xGGdK2"
		secret := "EI25dmZt7iiw_BiybAs7p2_6YZN198ULXg7T9M87hokdDFEM2PKeHrle2hCTJANexJQoEgrBy11Rc0Nb"
		client := paypal.NewClient(clientID, secret, paypal.APIBaseSandBox)
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
