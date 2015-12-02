package controller

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/sessions"
)

type CheckoutAPI struct {
	Store      *store.Module `inject:""`
	Components *components.Module `inject:""` 
	GCommerce  *gcommerce.Module `inject:""`
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

		// Check items against stored prices
		for id, item := range items {

			component_id := bson.ObjectIdHex(id)
			component, err :=  this.Components.Get(component_id)

			if err != nil {
				c.JSON(400, gin.H{"message": "Component in cart not found, hijacked.", "key": "invalid-component", "status": "error"})
				return
			}

			vendor, err := item.Attr("vendor")

			if err != nil {
				c.JSON(400, gin.H{"message": "Component in cart invalid, hijacked.", "key": "invalid-component-data", "status": "error"})
				return
			}

			price, err := component.GetVendorPrice(vendor.(string))

			if err != nil {
				c.JSON(400, gin.H{"message": "Vendor price does coulnt be verified.", "key": "invalid-vendor", "status": "error"})
				return
			}

			if item.Price != price {
				c.JSON(400, gin.H{"message": "Stored price and in-cart price have differences, perform check.", "key": "price-expired", "status": "error"})
				return
			}
		}

		customer := this.GCommerce.GetCustomerFromUser(userId)

		// Get a reference for the customer's address that will be used on the order
		address := customer.Address("mx", form.Delivery.State, form.Delivery.City, form.Delivery.Zipcode, form.Delivery.AddressLine1, form.Delivery.AddressLine2, form.Delivery.Extra)

		// Get a reference for the customer's new order
		order, err := customer.NewOrder(form.Gateway, form.Meta)

		if err != nil {
			c.JSON(400, gin.H{"message": err.Error(), "key": err.Error(), "status": "error"})
			return
		}

		for id, item := range items {

			meta := map[string]interface{}{
				"related": "components",
				"related_id": bson.ObjectIdHex(id),
				"cart": item.Attributes, 
			}

			order.Add(item.Name, "", "", item.Price, item.Quantity, meta)
		}

		// Setup shipping information
		order.Ship(120, "generic", address)

		err = order.Checkout()

		if err != nil {
			c.JSON(400, gin.H{"message": err.Error(), "key": err.Error(), "status": "error"})
			return
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
	Gateway  string       `json:"gateway" binding:"required"`
	Delivery DeliveryForm `json:"delivery" binding:"required"`	
	Meta     map[string]interface{} `json:"meta"`
}

type DeliveryForm struct {
	State   string `json:"state" binding:"required"`
	City    string `json:"city" binding:"required"`
	Zipcode string `json:"zipcode" binding:"required"`
	AddressLine1 string `json:"address_line1" binding:"required"`
	AddressLine2 string `json:"address_line2"`
	Extra        string `json:"extra"`
}