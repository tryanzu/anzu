package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

const ITEM_NOT_FOUND = "not-found"
const ITEM_NO_SELLER = "invalid-seller"
const ITEM_NOT_AVAILABLE = "cant-sell"
const ITEM_CHEAPER = "cheaper-now"
const ITEM_MORE_EXPENSIVE = "more-expensive"

type CheckoutAPI struct {
	Store      *store.Module      `inject:""`
	Components *components.Module `inject:""`
	GCommerce  *gcommerce.Module  `inject:""`
	Mail  *mail.Module   `inject:""`
	User  *user.Module   `inject:""`
}

func (this CheckoutAPI) Place(c *gin.Context) {

	var form CheckoutForm

	cartContainer := this.getCartObject(c)
	user := c.MustGet("user_id")
	userId := bson.ObjectIdHex(user.(string))

	if c.Bind(&form) == nil {

		items := cartContainer.GetContent()

		if len(items) == 0 {
			c.JSON(400, gin.H{"message": "No items in cart.", "status": "error"})
			return
		}

		shipping_cost := 0.0
		item_count := 0
		errors := make([]CheckoutError, 0)

		// Check items against stored prices
		for id, item := range items {

			component_id := bson.ObjectIdHex(id)
			component, err := this.Components.Get(component_id)

			if err != nil {

				errors = append(errors, CheckoutError{
					Type: ITEM_NOT_FOUND,
					Related: component_id,
				})

				// Remove from items
				cartContainer.Remove(id)
				delete(items, id)

				continue
			}

			vendor, err := item.Attr("vendor")

			if err != nil {

				errors = append(errors, CheckoutError{
					Type: ITEM_NO_SELLER,
					Related: component_id,
				})

				// Remove from items
				cartContainer.Remove(id)
				delete(items, id)

				continue
			}

			price, err := component.GetVendorPrice(vendor.(string))

			if err != nil {

				errors = append(errors, CheckoutError{
					Type: ITEM_NOT_AVAILABLE,
					Related: component_id,
				})

				// Remove from items
				cartContainer.Remove(id)
				delete(items, id)

				continue
			}

			if item.Price != price {

				if item.Price > price {

					errors = append(errors, CheckoutError{
						Type: ITEM_CHEAPER,
						Related: component_id,
						Meta: map[string]interface{}{
							"before": item.Price,
							"after": price,
						},
					})

					cartContainer.Update(id, item.Name, price, item.Quantity, item.Attributes)
					item.Price = price

					continue

				} else if item.Price < price {

					errors = append(errors, CheckoutError{
						Type: ITEM_MORE_EXPENSIVE,
						Related: component_id,
						Meta: map[string]interface{}{
							"before": item.Price,
							"after": price,
						},
					})

					cartContainer.Update(id, item.Name, price, item.Quantity, item.Attributes)
					item.Price = price

					continue
				}
			}

			if component.Type == "case" {

				shipping_cost = shipping_cost + 320.0

			} else {
				
				if item_count == 0 {

					shipping_cost = shipping_cost + 120.0
				} else {

					shipping_cost = shipping_cost + 60.0
				}

				item_count = item_count + 1
			}
		}

		if len(errors) > 0 {

			c.JSON(409, gin.H{"status": "error", "list": errors})
			return
		}

		customer := this.GCommerce.GetCustomerFromUser(userId)

		// Get a reference for the customer's address that will be used on the order
		address, err := customer.Address(form.ShipTo)

		if err != nil {
			c.JSON(400, gin.H{"message": "Invalid ship_to parameter.", "status": "error"})
			return
		}

		// Get a reference for the customer's new order
		order, err := customer.NewOrder(form.Gateway, form.Meta)

		if err != nil {
			c.JSON(400, gin.H{"message": err.Error(), "key": err.Error(), "status": "error"})
			return
		}

		for id, item := range items {

			meta := map[string]interface{}{
				"related":    "components",
				"related_id": bson.ObjectIdHex(id),
				"cart":       item.Attributes,
			}

			order.Add(item.Name, "", "", item.Price, item.Quantity, meta)
		}

		// Setup shipping information
		order.Ship(shipping_cost, "generic", address)

		// Match calculated total against frontend total
		total := order.GetTotal() 

		if total != form.Total {
			c.JSON(400, gin.H{"message": "Invalid total parameter.", "key": "bad-total", "status": "error"})
			return
		}

		err = order.Checkout()

		if err != nil {
			c.JSON(400, gin.H{"message": err.Error(), "key": err.Error(), "status": "error"})
			return
		}

		// After checkout procedures
		mailing := this.Mail
		{
			usr, err := this.User.Get(userId)

			if err != nil {
				panic(err)
			}

			var paymentType string

			if form.Gateway == "offline" {
				paymentType = "Transferencia o Deposito"
			} else if form.Gateway == "stripe" {
				paymentType = "Pago en linea"
			}

			compose := mail.Mail{
				Template: 252541,
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
					"total_products": order.Total - order.Shipping.Price,
					"total_shipping": order.Shipping.Price,
					"subtotal": order.Total,
					"total": order.Total,
				},
			}

			go mailing.Send(compose)
		}


		c.JSON(200, gin.H{"status": "okay"})
		return
	}

	c.JSON(400, gin.H{"message": "Invalid request, check order docs.", "status": "error"})
}

func (this CheckoutAPI) getCartObject(c *gin.Context) *cart.Cart {

	obj, err := cart.Boot(cart.GinGonicSession{sessions.Default(c)})

	if err != nil {
		panic(err)
	}

	return obj
}

type CheckoutForm struct {
	Gateway string                 `json:"gateway" binding:"required"`
	ShipTo  bson.ObjectId          `json:"ship_to" binding:"required"`
	Total   float64                `json:"total" binding:"required"`
	Meta    map[string]interface{} `json:"meta"`
}

type CheckoutError struct {
	Type    string `json:"type"`
	Related bson.ObjectId `json:"related_id"`
	Meta    map[string]interface{} `json:"data,omitempty"`
}