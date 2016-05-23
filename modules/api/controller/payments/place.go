package payments

import (
	"github.com/dustin/go-humanize"
	"github.com/fernandez14/spartangeek-blacker/modules/payments"
	"github.com/gin-gonic/gin"
	"github.com/leebenson/paypal"
	"gopkg.in/mgo.v2/bson"

	"time"
)

func (this API) Place(c *gin.Context) {

	var m PlacePayload

	if c.Bind(&m) == nil {

		client := this.GetPaypalClient()
		baseUrl, err := this.Config.String("application.siteUrl")

		if err != nil {
			panic("Could not get siteUrl from config.")
		}

		payment := paypal.Payment{
			Intent: "sale",
			Payer: &paypal.Payer{
				PaymentMethod: "paypal",
			},
			Transactions: []paypal.Transaction{
				{
					Amount: &paypal.Amount{
						Currency: "MXN",
						Total:    humanize.FormatFloat("###.##", m.Amount),
						/*Details: &paypal.Details{
							Shipping: "119.00",
							Subtotal: "116.00",
							Tax:      "3.00",
						},*/
					},
					Description: m.Description,
					ItemList: &paypal.ItemList{
						Items: []paypal.Item{
							{
								Quantity:    1,
								Name:        m.Description,
								Price:       humanize.FormatFloat("###.##", m.Amount),
								Currency:    "MXN",
								Description: m.Description,
								/*Tax:         "16.00",*/
							},
						},
					},
					SoftDescriptor: "SPARTANGEEK.COM",
				},
			},
			RedirectURLs: &paypal.RedirectURLs{
				CancelURL: baseUrl + "/donacion/error/",
				ReturnURL: baseUrl + "/donacion/exitosa/",
			},
		}

		dopayment, err := client.CreatePayment(payment)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "create-failed", "details": err})
			return
		}

		id := c.MustGet("user_id")
		user_id := bson.ObjectIdHex(id.(string))

		p := &payments.Payment{
			Type:      m.Type,
			Amount:    m.Amount,
			UserId:    user_id,
			Gateway:   "paypal",
			GatewayId: dopayment.ID,
			Meta:      dopayment,
			Status:    "created",
			Created:   time.Now(),
			Updated:   time.Now(),
		}

		err = this.Mongo.Save(p)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "save-failed", "details": err})
			return
		}

		var approval string

		for _, l := range dopayment.Links {
			if l.Rel == "approval_url" {
				approval = l.Href
			}
		}

		if approval == "" {
			c.JSON(400, gin.H{"status": "error", "error": "approval-failed", "details": "Blacker could not get approval URL from PayPal response."})
			return
		}

		c.JSON(200, gin.H{"status": "okay", "response": gin.H{"approval_url": approval}})
	}
}
