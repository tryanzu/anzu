package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/go-siftscience"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

type CartAPI struct {
	Components *components.Module `inject:""`
}

// Get Cart items
func (this CartAPI) Get(c *gin.Context) {

	var items []CartComponentItem

	// Initialize cart library
	err := this.getCart(c).Bind(&items)

	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
	} else {

		if len(items) > 0 {
			c.JSON(200, items)
		} else {
			c.JSON(200, make([]string, 0))
		}
	}
}

// Add Cart item from component id
func (this CartAPI) Add(c *gin.Context) {

	var form CartAddForm

	if c.BindJSON(&form) == nil {

		id := form.Id
		session_id := c.MustGet("session_id").(string)

		if !bson.IsObjectIdHex(id) {
			c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
			return
		}

		// Get the component and its data
		component_id := bson.ObjectIdHex(id)
		component, err := this.Components.Get(component_id)

		if err != nil {
			c.JSON(404, gin.H{"message": "Invalid request, component not found.", "status": "error"})
			return
		}

		// Initialize cart library
		container := this.getCart(c)
		{
			var items []CartComponentItem
			var cartItem CartComponentItem

			err := container.Bind(&items)

			if err != nil {
				c.JSON(500, gin.H{"status": "error", "message": err.Error()})
				return
			}

			price, err := component.GetVendorPrice(form.Vendor)

			if err != nil {
				c.JSON(400, gin.H{"message": "Invalid vendor, check id.", "status": "error"})
				return
			}

			attrs := map[string]interface{}{
				"vendor": form.Vendor,
			}

			exists := false

			for index, item := range items {

				if item.Id == component.Id.Hex() {
					exists = true
					items[index].IncQuantity(1)
					cartItem = items[index]
					break
				}
			}

			user_id_i, signed_in := c.Get("user_id")

			if !exists {

				base := cart.CartItem{
					Id: component.Id.Hex(),
					Name: component.Name,
					Price: price,
					Quantity: 1,
					Attributes: attrs,
				}

				item := CartComponentItem{base, component.FullName, component.Image, component.Slug, component.Type}
				items = append(items, item)
				cartItem = item
			}

			go func() {

				data := map[string]interface{}{
					"$session_id": session_id,
					"$item": this.generateSiftItem(cartItem, component),
				}

				if signed_in {
					data["$user_id"] = user_id_i.(string)
				}

				err := gosift.Track("$add_item_to_cart", data)

				if err != nil {
					panic(err)
				}
			}()

			err = container.Save(items)

			if err != nil {
				c.JSON(500, gin.H{"status": "error", "message": err.Error()})
				return
			}
		}

		c.JSON(200, gin.H{"status": "okay"})
	}
}

// Delete Item from cart
func (this CartAPI) Delete(c *gin.Context) {

	id := c.Param("id")
	session_id := c.MustGet("session_id").(string)

	if !bson.IsObjectIdHex(id) {
		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	container := this.getCart(c)
	{
		var items []CartComponentItem
		var cartItem CartComponentItem
		var matched bool = false

		err := container.Bind(&items)

		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": err.Error()})
			return
		}

		for i, item := range items {

			if item.Id == id {

				items[i].Quantity = items[i].Quantity - 1
				cartItem = items[i]

				if items[i].Quantity < 1 {
					items = append(items[:i], items[i+1:]...)
				}

				matched = true

				break
			}
		}

		if matched {

			user_id_i, signed_in := c.Get("user_id")

			go func() {

				cartItem.Quantity = 1
				component_id := bson.ObjectIdHex(cartItem.Id)
				component, err := this.Components.Get(component_id)

				if err == nil {

					data := map[string]interface{}{
						"$session_id": session_id,
						"$item": this.generateSiftItem(cartItem, component),
					}

					if signed_in {
						data["$user_id"] = user_id_i.(string)
					}

					err := gosift.Track("$remove_item_from_cart", data)

					if err != nil {
						panic(err)
					}
				}
			}()
		}

		container.Save(items)

		c.JSON(200, gin.H{"status": "okay"})
	}
}

func (this CartAPI) getCart(c *gin.Context) *cart.Cart {

	obj, err := cart.Boot(cart.GinGonicSession{sessions.Default(c)})

	if err != nil {
		panic(err)
	}

	return obj
}

func (this CartAPI) generateSiftItem(c CartComponentItem, component *components.ComponentModel) map[string]interface{} {

	micros := int64((c.Price * 100) * 10000)

	data := map[string]interface{}{
		"$item_id": c.Id,
		"$product_title": c.FullName,
		"$price": micros,
		"$currency_code": "MXN",
		"$brand": component.Manufacturer,
		"$manufacturer": component.Manufacturer,
		"$category": component.Type,
		"$quantity": c.Quantity,
	}

	return data
}

type CartAddForm struct {
	Type   string `json:"type" binding:"required"`
	Id     string `json:"id" binding:"required"`
	Vendor string `json:"vendor"`
}

type CartComponentItem struct {
	cart.CartItem
	FullName string `json:"full_name"`
	Image    string `json:"image"`
	Slug     string `json:"slug"`
	Type     string `json:"type"`
}
