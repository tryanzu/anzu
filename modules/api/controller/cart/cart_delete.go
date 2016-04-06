package cart

import (
	"github.com/fernandez14/go-siftscience"
	"github.com/fernandez14/spartangeek-blacker/modules/cart"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func (this API) Delete(c *gin.Context) {

	id := c.Param("id")
	session_id := c.MustGet("session_id").(string)

	if !bson.IsObjectIdHex(id) {
		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	container := this.getCart(c)
	{
		var items []cart.CartItem
		var matchedItem cart.CartItem
		var matched bool = false

		err := container.Bind(&items)

		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": err.Error()})
			return
		}

		for i, item := range items {

			if item.Id == id {

				items[i].Quantity = items[i].Quantity - 1
				matchedItem = items[i]

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

				matchedItem.Quantity = 1
				products := this.GCommerce.Products()
				product_id := bson.ObjectIdHex(matchedItem.Id)
				product, err := products.GetById(product_id)

				if err == nil {

					data := map[string]interface{}{
						"$session_id": session_id,
						"$item":       this.generateSiftItem(product),
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
