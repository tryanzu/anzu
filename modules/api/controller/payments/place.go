package payments

import (
	"github.com/fernandez14/spartangeek-blacker/modules/payments"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Place(c *gin.Context) {

	var m PlacePayload

	if c.Bind(&m) == nil {

		// Create a payment using paypal gateway
		gateway := this.GetPaypalGateway(map[string]interface{}{})
		transaction := this.Payments.Create(gateway)

		// Get user from session middleware
		id := c.MustGet("user_id")
		user_id := bson.ObjectIdHex(id.(string))

		var products []payments.Product

		products = append(products, &payments.DigitalProduct{
			Name:        m.Description,
			Description: m.Description,
			Quantity:    1,
			Price:       m.Amount,
			Currency:    "MXN",
		})

		transaction.SetUser(user_id)
		transaction.SetIntent(payments.DONATION)
		transaction.SetProducts(products)

		payment, res, err := transaction.Purchase()

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "create-failed", "details": err})
			return
		}

		c.JSON(200, gin.H{"status": "okay", "payment_status": payment.Status, "response": res})
	}
}
