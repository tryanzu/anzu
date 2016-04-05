package cart

// Add Cart item from component id
func (this API) Add(c *gin.Context) {

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