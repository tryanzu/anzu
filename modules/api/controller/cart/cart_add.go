package cart

import (
	"github.com/fernandez14/go-siftscience"
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

// Add Cart item from type & id
func (this API) Add(c *gin.Context) {

	var form CartAddForm

	if c.BindJSON(&form) == nil {

		id := form.Id
		session_id := c.MustGet("session_id").(string)

		if !bson.IsObjectIdHex(id) {
			c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
			return
		}

		products := this.GCommerce.Products()
		product_id := bson.ObjectIdHex(id)
		product, err := products.GetById(product_id)

		if err != nil {
			c.JSON(404, gin.H{"message": "Invalid request, product not found.", "status": "error"})
			return
		}

		// Initialize cart library
		container := this.getCart(c)
		{
			var items []cart.CartItem

			err := container.Bind(&items)

			if err != nil {
				c.JSON(500, gin.H{"status": "error", "message": err.Error()})
				return
			}

			exists := false

			for index, item := range items {

				if item.Id == product.Id.Hex() {
					exists = true
					items[index].IncQuantity(1)
					break
				}
			}

			user_id_i, signed_in := c.Get("user_id")

			if !exists {

				item := cart.CartItem{
					Id:          product.Id.Hex(),
					Name:        product.Name,
					Description: product.Description,
					Image:       product.Image,
					Price:       product.Price,
					Type:        product.Type,
					Quantity:    1,
					Attributes:  map[string]interface{}{},
				}

				items = append(items, item)
			}

			go func() {

				data := map[string]interface{}{
					"$session_id": session_id,
					"$item":       this.generateSiftItem(product),
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
		return
	}

	c.JSON(400, gin.H{"status": "error", "message": "Malformed request."})
}
