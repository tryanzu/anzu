package controller 

import (
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"strconv"
)

type OrdersAPI struct {
	GCommerce  *gcommerce.Module  `inject:""`
	Mail  *mail.Module   `inject:""`
	User  *user.Module   `inject:""`
}

func (this OrdersAPI) Get(c *gin.Context) {

	var offset int = 0
	var limit  int = 20

	l := c.Query("limit")
	o := c.Query("offset")
	search := c.Query("search")

	if l != "" {
		cl, err := strconv.Atoi(l)

		if err == nil && cl > 0 {
			limit = cl 
		}
	}

	if o != "" {
		co, err := strconv.Atoi(o)

		if err == nil && co > 0 {
			offset = co 
		}
	}

	meta := bson.M{}

	if search != "" {

		meta = bson.M{
			"$or": []bson.M{
				{
					"reference": bson.M{
						"$regex": search,
					},
				},
				{
					"$text": bson.M{
						"$search": search,
					},
				},
			},
		}
	}

	orders := this.GCommerce.Get(meta, limit, offset)

	if len(orders) == 0 {
		c.JSON(200, make([]string, 0))
		return
	}

	c.JSON(200, orders)
}

func (this OrdersAPI) SendOrderConfirmation(c *gin.Context) {

	order_id := c.Param("id")

	if bson.IsObjectIdHex(order_id) == false {
		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(order_id)
	order, err := this.GCommerce.One(bson.M{"_id": id})

	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid request, order not found.", "status": "error"})
		return
	}

	customer, err := this.GCommerce.GetCustomer(order.UserId)

	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid request, order customer not found.", "status": "error"})
		return
	}

	usr, err := this.User.Get(customer.UserId)

	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid request, order user not found.", "status": "error"})
		return
	}

	address_id, exists := order.Shipping.Meta["related_id"]

	if exists {

		address, err := customer.Address(address_id.(bson.ObjectId))

		if err != nil {
			c.JSON(404, gin.H{"message": "Invalid request, order address not found.", "status": "error"})
			return
		}

		mailing := this.Mail
		{	
			var paymentType string
			var template int

			if order.Gateway == "offline" {
				paymentType = "Transferencia o Deposito"
				template = 324721
			} else if order.Gateway == "stripe" {
				paymentType = "Pago en linea"
				template = 252541
			}

			compose := mail.Mail{
				Template: template,
				Recipient: []mail.MailRecipient{
					{
						Name:  usr.Name(),
						Email: usr.Email(),
					},
				},
				Variables: map[string]interface{}{
					"name": usr.Name(),
					"payment": paymentType,
					"line1": address.Line1(),
					"line2": address.Line2(),
					"line3": address.Extra(),
					"total_products": order.GetOriginalTotal() - order.Shipping.OPrice,
					"total_shipping": order.Shipping.OPrice,
					"subtotal": order.GetOriginalTotal(),
					"commision": order.GetGatewayCommision(),
					"total": order.Total,
					"items": order.Items,
					"reference": order.Reference,
				},
			}

			go mailing.Send(compose)
		}
	}
}

func (this OrdersAPI) ChangeStatus(c *gin.Context) {

	var form ComponentStatusForm
	
	order_id := c.Param("id")
	
	if bson.IsObjectIdHex(order_id) == false {
		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(order_id)
	order, err := this.GCommerce.One(bson.M{"_id": id})

	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid request, order not found.", "status": "error"})
		return
	}

	if c.Bind(&form) == nil {
		
	 	if gcommerce.ValidStatus(form.Name) {
	 		
	 		order.ChangeStatus(form.Name)
	 		
	 		c.JSON(200, gin.H{"status": "okay"})
	 		return
	 	}
	}
	
	c.JSON(400, gin.H{"message": "Invalid request, status not valid.", "status": "error"})
	return
}

type ComponentStatusForm struct {
	Name string `json:"status" binding:"required"`
}