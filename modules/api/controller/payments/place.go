package payments

import (
	"github.com/dustin/go-humanize"
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/queue"
	"github.com/gin-gonic/gin"
	"github.com/leebenson/paypal"
	"gopkg.in/mgo.v2/bson"

	"time"
)

func (this API) Place(c *gin.Context) {

	var m PlacePayload

	if c.Bind(&m) == nil {

		database := this.Mongo.Database
		clientID := "ASkcCWSuhGXVNim-dfAJ9Zlxk41iLceeLLSv7_dvBZY-Dob1sGBVFaMgIUKaOyHb9TmjWXgV83xGGdK2"
		secret := "EI25dmZt7iiw_BiybAs7p2_6YZN198ULXg7T9M87hokdDFEM2PKeHrle2hCTJANexJQoEgrBy11Rc0Nb"
		client := paypal.NewClient(clientID, secret, paypal.APIBaseSandBox)

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
				CancelURL: "https://spartangeek.com/donacion/fallida/",
				ReturnURL: "https://spartangeek.com/donacion/exitosa/",
			},
		}

		dopayment, err := client.CreatePayment(payment)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": "create-failed", "details": err})
			return
		}

		p := Payment{
			Type:      m.Type,
			Amount:    m.Amount,
			Gateway:   "paypal",
			GatewayId: dopayment.ID,
			Meta:      dopayment,
			Status:    "created",
			Created:   time.Now(),
			Updated:   time.Now(),
		}

		err = database.C("payments").Insert(p)

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

		c.JSON(200, gin.H{"status": "okay", "response": gin.M{"approval_url": approval}})
	}
}

type Payment struct {
	Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Type      string        `bson:"type" json:"type"`
	Amount    float64       `bson:"amount" json:"amount"`
	Gateway   string        `bson:"gateway" json:"gateway"`
	GatewayId string        `bson:"gateway_id" json:"gateway_id"`
	Meta      interface{}   `bson:"gateway_response,omitempty" json:"gateway_response,omitempty"` // TODO - Move it to another collection
	Status    string        `bson:"status" json:"status"`
	Created   time.Time     `bson:"created_at" json:"created_at"`
	Updated   time.Time     `bson:"updated_at" json:"updated_at"`
}
