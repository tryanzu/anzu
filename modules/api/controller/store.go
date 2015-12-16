package controller

import (
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
)

type StoreAPI struct {
	Store *store.Module `inject:""`
}

// Place an order (public endpoint)
func (self StoreAPI) PlaceOrder(c *gin.Context) {

	var form OrderForm

	var ip string = c.ClientIP()

	if c.BindJSON(&form) == nil {

		order := store.OrderModel{
			User: store.OrderUserModel{
				Name:  form.User.Name,
				Email: form.User.Email,
				Phone: form.User.Phone,
				Ip: ip,
			},
			Content:  form.Content,
			Budget:   form.Budget,
			Currency: "MXN",
			State:    form.State,
			Games:    form.Games,
			Extra:    form.Extra,
			Usage:    form.Usage,
			Unreaded: true,
			BuyDelay: form.BuyDelay,
		}

		self.Store.CreateOrder(order)

		c.JSON(200, gin.H{"status": "okay", "signed": ip})
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

// REST handler for getting one order
func (self StoreAPI) One(c *gin.Context) {

	id := c.Param("id")

	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"status": "error", "message": "Can't perform action. Invalid id."})
		return
	}

	order, err := self.Store.Order(bson.ObjectIdHex(id))

	if err != nil {

		c.JSON(404, gin.H{"status": "error", "message": "Order not found."})
		return
	}

	// Mark as readed
	order.Touch()

	// Load assets
	order.LoadAssets()

	data := order.Data()
	data.RelatedUsers = order.MatchUsers()


	c.JSON(200, data)
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

			order.PushAnswer(form.Content, form.Type)

			c.JSON(200, gin.H{"status": "okay"})
		} else {

			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		}
	}
}

// Push tag to an order
func (self *StoreAPI) Tag(c *gin.Context) {

	var form OrderTagForm

	order_id := c.Param("id")

	if bson.IsObjectIdHex(order_id) == false {

		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(order_id)

	if c.BindJSON(&form) == nil {

		order, err := self.Store.Order(id)

		if err == nil {

			order.PushTag(form.Name)

			c.JSON(200, gin.H{"status": "okay"})
		} else {

			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		}
	}
}

// Update order stage
func (self *StoreAPI) Stage(c *gin.Context) {

	var form OrderStageForm

	order_id := c.Param("id")

	if bson.IsObjectIdHex(order_id) == false {

		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(order_id)

	if c.BindJSON(&form) == nil {

		order, err := self.Store.Order(id)

		if err == nil {

			order.Stage(form.Name)

			c.JSON(200, gin.H{"status": "okay"})
		} else {

			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		}
	}
}

// Push activity for the order
func (self *StoreAPI) Activity(c *gin.Context) {

	var form OrderActivityForm

	order_id := c.Param("id")

	if bson.IsObjectIdHex(order_id) == false {

		c.JSON(400, gin.H{"message": "Invalid request, id not valid.", "status": "error"})
		return
	}

	id := bson.ObjectIdHex(order_id)

	if c.BindJSON(&form) == nil {

		due_at, err := time.Parse("2006-01-02 15:04", form.Due)

		if err != nil {

			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		} else {

			order, err := self.Store.Order(id)

			if err == nil {

				order.PushActivity(form.Name, form.Description, due_at)

				c.JSON(200, gin.H{"status": "okay"})
			} else {

				c.JSON(400, gin.H{"status": "error", "message": err.Error()})
			}
		}
	}
}

type OrderForm struct {
	User     OrderUserForm `json:"user" binding:"required"`
	Content  string        `json:"content"`
	Budget   int           `json:"budget" binding:"required"`
	BuyDelay int           `json:"buydelay" binding:"required"`
	State    string        `json:"estado" binding:"required"`
	Usage    string		   `json:"usage"`
	Games    []string      `json:"games"`
	Extra    []string      `json:"extra"`
}

type OrderUserForm struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type OrderTagForm struct {
	Name  string `json:"name" binding:"required"`
}

type OrderStageForm struct {
	Name  string `json:"name" binding:"required"`
}

type OrderActivityForm struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	Due          string `json:"due_at" binding:"required"`
}

type OrderAnswerForm struct {
	Content string `json:"content" binding:"required"`
	Type    string `json:"type" binding:"required"`
}
