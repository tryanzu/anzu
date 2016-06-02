package payments

import (
	"github.com/fernandez14/spartangeek-blacker/modules/payments"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	//"strconv"
	//"time"
)

func (this API) PaypalExecute(c *gin.Context) {

	var m ExecutePayload

	if c.Bind(&m) == nil {

		payment, err := this.Payments.Get(bson.M{"gateway": "paypal", "gateway_id": m.PaymentId})

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "not-found", "payment_id": m.PaymentId, "details": err})
			return
		}

		if payment.Status != payments.PAYMENT_AWAITING {
			c.JSON(400, gin.H{"status": "error", "error": "invalid-action", "details": "Can't execute this payment."})
			return
		}

		res, err := payment.CompletePurchase(map[string]interface{}{
			"payer_id": m.PayerId,
		})

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "execution-failed", "details": err})
			return
		}

		if payment.Type == "sale" && payment.Related == "order" && payment.RelatedId.Valid() {

			order, err := this.GCommerce.One(bson.M{"_id": payment.RelatedId})

			if err == nil {
				order.ChangeStatus("confirmed")
			}
		}

		id := c.MustGet("user_id")
		user_id := bson.ObjectIdHex(id.(string))
		usr := this.Gaming.Get(user_id)

		if err != nil {
			panic(err)
		}

		total := payment.Amount

		// Gift user a badge - TODO: find better way for this matter
		if total >= 50 && total < 200 {
			usr.AcquireBadge(bson.ObjectIdHex("55f0a94fbe6bceb9b3762c6c"), false)
		} else if total >= 200 && total < 1000 {
			usr.AcquireBadge(bson.ObjectIdHex("55f0a950be6bceb9b3762c6d"), false)
		} else if total > 1000 {
			usr.AcquireBadge(bson.ObjectIdHex("55f0a951be6bceb9b3762c6e"), false)
		}

		c.JSON(200, gin.H{"status": "okay", "response": res})
	}
}
