package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

type StoreAPI struct {
	Store *store.Module `inject:""`
}

// Place an order (public endpoint)
func (self StoreAPI) PlaceOrder(c *gin.Context) {

	var form OrderForm

	if c.BindJSON(&form) == nil {

		order := store.OrderModel{
			User: store.OrderUserModel{
				Name:  form.User.Name,
				Email: form.User.Email,
				Phone: form.User.Phone,
			},
			Content:  form.Content,
			Budget:   form.Budget,
			Currency: form.Currency,
			Games:    form.Games,
			Extra:    form.Extra,
			BuyDelay: form.BuyDelay,
		}

		self.Store.CreateOrder(order)

		c.JSON(200, gin.H{"status": "okay"})
	}
}

// Get all orders sorted by convenience
func (self StoreAPI) Orders(c *gin.Context) {

	// Get parameters and convert them to needed types
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if err != nil {
		offset = 0
	}

	orders := self.Store.GetSortedOrders(limit, offset)

	c.JSON(200, orders)
}

// Use one of the predefined answers to answer an order
func (self StoreAPI) FastAnswer(c *gin.Context) {

}

// Answer with text to an order
func (self *StoreAPI) Answer(c *gin.Context) {

	var form OrderAnswerForm

	order_id := c.Param("id")

	if bson.IsObjectIdHex(order_id) == false {

		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(order_id)

	if c.BindJSON(&form) == nil {

		order, err := self.Store.Order(id)

		if err == nil {

			order.PushAnswer(form.Content)

			c.JSON(200, gin.H{"status": "okay"})
		} else {

			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		}
	}
}

type OrderForm struct {
	User     OrderUserForm `json:"user" binding:"required"`
	Content  string        `json:"content" binding:"required"`
	Budget   int           `json:"budget" binding:"required"`
	Currency string        `json:"currency" binding:"required"`
	BuyDelay int           `json:"buydelay" binding:"required"`
	Games    []string      `json:"games"`
	Extra    []string      `json:"extra"`
}

type OrderUserForm struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type OrderAnswerForm struct {
	Content string `json:"content" binding:"required"`
}
