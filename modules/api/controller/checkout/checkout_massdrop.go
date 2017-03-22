package checkout

import (
	"github.com/dustin/go-humanize"
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Massdrop(c *gin.Context) {

	var form MassdropForm

	id := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(id.(string))
	session_id := c.MustGet("session_id").(string)

	if c.Bind(&form) == nil {

		products := this.GCommerce.Products()
		product, err := products.GetById(form.ProductId)

		if err != nil {
			c.JSON(400, gin.H{"message": "Invalid product_id, can't find product", "error": "not-found", "status": "error"})
			return
		}

		// Load Massdrop information (if exists)
		product.InitializeMassdrop()

		if product.Massdrop == nil {
			c.JSON(400, gin.H{"message": "Invalid product_id, can't checkout massdrop for product", "error": "invalid-massdrop", "status": "error"})
			return
		}

		if product.Massdrop.Reserve <= 0 {
			c.JSON(400, gin.H{"message": "Invalid product_id, can't checkout massdrop for product", "error": "invalid-massdrop-reserve", "status": "error"})
			return
		}

		if product.Massdrop.Active == false {
			c.JSON(400, gin.H{"message": "Invalid product_id, can't checkout massdrop for product", "error": "massdrop-finished", "status": "error"})
			return
		}

		if form.Quantity <= 0 {
			c.JSON(400, gin.H{"message": "Invalid quantity, can't checkout massdrop for product", "error": "invalid-quantity", "status": "error"})
			return
		}

		customer := this.GCommerce.GetCustomerFromUser(user_id)

		// Get a reference for the customer's new order
		meta := form.Meta
		meta["session_id"] = session_id

		res, order, transaction, err := customer.MassdropTransaction(product, form.Quantity, form.Gateway, meta)

		if err != nil {
			c.JSON(400, gin.H{"message": err.Error(), "error": err.Error(), "status": "error"})
			return
		}

		if form.Gateway != "paypal" {

			// After checkout procedures
			mailing := deps.Container.Mailer()
			{
				usr, err := this.User.Get(user_id)

				if err != nil {
					panic(err)
				}

				var template int = 549841

				if form.Gateway == "offline" {
					template = 549941
				}

				compose := mail.Mail{
					mail.MailBase{
						FromName:  "Spartan Geek",
						FromEmail: "pedidos@spartangeek.com",
						Recipient: []mail.MailRecipient{
							{
								Name:  usr.Name(),
								Email: usr.Email(),
							},
							{
								Name:  "Equipo Spartan Geek",
								Email: "pedidos@spartangeek.com",
							},
						},
						Variables: map[string]interface{}{
							"name":      usr.Name(),
							"reference": order.Reference,
							"price":     product.Massdrop.Reserve,
							"slug":      product.Slug,
							"pname":     product.Name,
							"quantity":  form.Quantity,
							"total":     humanize.FormatFloat("#,###.##", product.Massdrop.Reserve*float64(form.Quantity)),
						},
					},
					template,
				}

				go mailing.Send(compose)
			}
		}

		c.JSON(200, gin.H{"status": "okay", "response": res, "transaction_id": transaction.Id})
		return
	}

	c.JSON(400, gin.H{"message": "Malformed request, check checkout docs.", "status": "error"})
}
