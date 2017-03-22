package checkout

import (
	"github.com/fernandez14/spartangeek-blacker/deps"
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Place(c *gin.Context) {

	var form CheckoutForm

	cartContainer := this.getCartObject(c)
	id := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(id.(string))
	session_id := c.MustGet("session_id").(string)

	if c.Bind(&form) == nil {

		var items []cart.CartItem

		// Initialize cart library
		err := cartContainer.Bind(&items)

		customer := this.GCommerce.GetCustomerFromUser(user_id)
		static, serr := customer.GetCart()

		if len(static) > 0 && serr == nil {
			for _, item := range static {
				items = append(items, item.Item)
			}
		}

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
		massdrop_enabled := false
		errors := make([]CheckoutError, 0)

		products := this.GCommerce.Products()
		plist := map[string]*gcommerce.Product{}

		// Check items against stored prices
		for index, item := range items {

			id := item.Id
			product_id := bson.ObjectIdHex(id)
			product, err := products.GetById(product_id)

			if err != nil {

				errors = append(errors, CheckoutError{
					Type:    ITEM_NOT_FOUND,
					Related: product_id,
				})

				// Remove from items
				items = append(items[:index], items[index+1:]...)

				continue
			}

			plist[id] = product

			if item.GetPrice() != product.Price {

				if item.GetPrice() > product.Price {

					errors = append(errors, CheckoutError{
						Type:    ITEM_CHEAPER,
						Related: product_id,
						Meta: map[string]interface{}{
							"before": item.GetPrice(),
							"after":  product.Price,
						},
					})

					item.SetPrice(product.Price)

					continue

				} else if item.GetPrice() < product.Price {

					errors = append(errors, CheckoutError{
						Type:    ITEM_MORE_EXPENSIVE,
						Related: product_id,
						Meta: map[string]interface{}{
							"before": item.GetPrice(),
							"after":  product.Price,
						},
					})

					item.SetPrice(product.Price)

					continue
				}
			}

			if product.Shipping > 0 {

				shipping_cost = shipping_cost + (product.Shipping * float64(item.GetQuantity()))

			} else if product.Category == "case" {

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

			if me, exists := item.Attributes["massdrop"].(bool); exists && me {
				massdrop_enabled = true
			}
		}

		if len(errors) > 0 {
			cartContainer.Save(items)

			c.JSON(409, gin.H{"status": "error", "list": errors})
			return
		}

		shipMethod := form.ShipMethod

		if shipMethod != "pickup" && shipMethod != "generic" {
			c.JSON(400, gin.H{"message": "Invalid ship_method parameter.", "status": "error"})
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

		if rfc, exists := meta["rfc"].(string); exists {
			if fiscal_ref, fe := meta["razon_social"].(string); fe {
				customer.UpdateTaxData(rfc, fiscal_ref)
			}
		}

		for _, item := range items {

			id := item.Id
			meta := map[string]interface{}{
				"related":    "product",
				"related_id": bson.ObjectIdHex(id),
				"cart":       item,
			}

			order.Add(item.Name, item.Description, item.Image, item.GetPrice(), item.GetQuantity(), meta)
		}

		if shipMethod == "generic" {

			if !form.ShipTo.Valid() {
				c.JSON(400, gin.H{"message": "Invalid ship_to id parameter.", "status": "error"})
				return
			}

			// Get a reference for the customer's address that will be used on the order
			address, err := customer.Address(form.ShipTo)

			if err != nil {
				c.JSON(400, gin.H{"message": "Invalid ship_to parameter.", "status": "error"})
				return
			}

			// Setup shipping information
			order.Ship(shipping_cost, "generic", address)

		} else if shipMethod == "pickup" {

			if !massdrop_enabled {
				c.JSON(400, gin.H{"message": "Pickup shipping not enabled.", "status": "error"})
				return
			}

			order.Ship(0, "pickup", nil)
		}

		// Match calculated total against frontend total
		total := order.GetTotal()

		if total != form.Total {
			c.JSON(400, gin.H{"message": "Invalid total parameter.", "key": "bad-total", "status": "error", "shipping": order.Shipping.Price, "total": total})
			return
		}

		err = order.Save()

		if err != nil {
			c.JSON(400, gin.H{"message": err.Error(), "key": err.Error(), "status": "error"})
			return
		}

		res, err := order.Checkout()

		if err != nil {
			c.JSON(400, gin.H{"message": err.Error(), "key": err.Error(), "status": "error"})
			return
		}

		// After checkout procedures
		mailing := deps.Container.Mailer()
		{
			usr, err := this.User.Get(user_id)

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

			vars := map[string]interface{}{
				"name":           usr.Name(),
				"payment":        paymentType,
				"total_products": order.GetOriginalTotal() - order.Shipping.OPrice,
				"total_shipping": order.Shipping.OPrice,
				"subtotal":       order.GetOriginalTotal(),
				"commision":      order.GetGatewayCommision(),
				"total":          order.Total,
				"items":          order.Items,
				"reference":      order.Reference,
			}

			if order.CustomerAdress != nil {
				vars["line1"] = order.CustomerAdress.Line1()
				vars["line2"] = order.CustomerAdress.Line2()
				vars["line3"] = order.CustomerAdress.Extra()
			}

			compose := mail.Mail{
				Template:  template,
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
				Variables: vars,
			}

			go mailing.Send(compose)

			// Clean up cart items
			cartContainer.Save(make([]cart.CartItem, 0))
			customer.CleanCart()
		}

		c.JSON(200, gin.H{"status": "okay", "response": res})
		return
	}

	c.JSON(400, gin.H{"message": "Malformed request, check order docs.", "status": "error"})
}
