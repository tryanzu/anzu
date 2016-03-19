package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/queue"
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
	usr := c.MustGet("user_id")
	userId := bson.ObjectIdHex(usr.(string))
	session_id := c.MustGet("session_id").(string)

	if c.Bind(&form) == nil {

		var items []CartComponentItem

		// Initialize cart library
		err := cartContainer.Bind(&items)

		if err != nil || len(items) == 0 {

			if err != nil {
				c.JSON(500, gin.H{"message": err.Error(), "status": "error"})
			} else {
				c.JSON(400, gin.H{"message": "No items in cart.", "status": "error"})
			}

			return
		}

		shipping_cost := 0.0
		item_count := 0
		errors := make([]CheckoutError, 0)

		clist := map[string]*components.ComponentModel{}

		// Check items against stored prices
		for index, item := range items {

			id := item.Id
			component_id := bson.ObjectIdHex(id)
			component, err := this.Components.Get(component_id)

			if err != nil {

				errors = append(errors, CheckoutError{
					Type: ITEM_NOT_FOUND,
					Related: component_id,
				})

				// Remove from items
				items = append(items[:index], items[index+1:]...)

				continue
			}

			clist[id] = component
			vendor, err := item.Attr("vendor")

			if err != nil {

				errors = append(errors, CheckoutError{
					Type: ITEM_NO_SELLER,
					Related: component_id,
				})

				// Remove from items
				items = append(items[:index], items[index+1:]...)

				continue
			}

			price, err := component.GetVendorPrice(vendor.(string))

			if err != nil {

				errors = append(errors, CheckoutError{
					Type: ITEM_NOT_AVAILABLE,
					Related: component_id,
				})

				// Remove from items
				items = append(items[:index], items[index+1:]...)

				continue
			}

			if item.GetPrice() != price {

				if item.GetPrice() > price {

					errors = append(errors, CheckoutError{
						Type: ITEM_CHEAPER,
						Related: component_id,
						Meta: map[string]interface{}{
							"before": item.GetPrice(),
							"after": price,
						},
					})

					item.SetPrice(price)

					continue

				} else if item.GetPrice() < price {

					errors = append(errors, CheckoutError{
						Type: ITEM_MORE_EXPENSIVE,
						Related: component_id,
						Meta: map[string]interface{}{
							"before": item.GetPrice(),
							"after": price,
						},
					})

					item.SetPrice(price)

					continue
				}
			}

			if component.Type == "case" {

				shipping_cost = shipping_cost + (320.0 * float64(item.GetQuantity()))

			} else {

				for i := 0; i < item.GetQuantity(); i++ {

					if item_count == 0 {

						shipping_cost = shipping_cost + 139.0
					} else {

						shipping_cost = shipping_cost + 60.0
					}

					item_count = item_count + 1
				}
			}
		}

		if len(errors) > 0 {

			cartContainer.Save(items)

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
		meta := form.Meta
		meta["session_id"] = session_id

		order, err := customer.NewOrder(form.Gateway, meta)

		if err != nil {
			c.JSON(400, gin.H{"message": err.Error(), "key": err.Error(), "status": "error", "order_id": order.Id})
			return
		}

		for _, item := range items {

			id := item.Id
			meta := map[string]interface{}{
				"related":    "components",
				"related_id": bson.ObjectIdHex(id),
				"cart":       item,
			}

			description := ""
			image := ""

			if c, exists := clist[id]; exists {

				description = c.Manufacturer + " / " +  c.PartNumber
				image = c.Image
			}

			name := item.FullName

			if name == "" {
				name = item.GetName()
			}

			order.Add(name, description, image, item.GetPrice(), item.GetQuantity(), meta)
		}

		// Setup shipping information
		order.Ship(shipping_cost, "generic", address)

		// Match calculated total against frontend total
		total := order.GetTotal()

		if total != form.Total {
			c.JSON(400, gin.H{"message": "Invalid total parameter.", "key": "bad-total", "status": "error", "shipping": shipping_cost, "total": total})
			return
		}

		err = order.Save()

		if err != nil {
			c.JSON(400, gin.H{"message": err.Error(), "key": err.Error(), "status": "error"})
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
			var template int

			if form.Gateway == "offline" {
				paymentType = "Transferencia o Deposito"
				template = 324721
			} else if form.Gateway == "stripe" {
				paymentType = "Pago en linea"
				template = 252541
			}

			compose := mail.Mail{
				Template: template,
				FromName: "Spartan Geek",
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

			go func(id bson.ObjectId) {

				err := queue.PushWDelay("gcommerce", "payment-reminder", map[string]interface{}{"id": id.Hex()}, 3600*24*2)

				if err != nil {
					panic(err)
				}

			}(order.Id)

			// Clean up cart items
			cartContainer.Save(make([]CartComponentItem, 0))
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
	Order   bson.ObjectId          `json:"order_id"`
	Total   float64                `json:"total" binding:"required"`
	Meta    map[string]interface{} `json:"meta"`
}

type CheckoutError struct {
	Type    string `json:"type"`
	Related bson.ObjectId `json:"related_id"`
	Meta    map[string]interface{} `json:"data,omitempty"`
}
