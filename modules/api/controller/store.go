package controller 

import (
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/gin-gonic/gin"
)

type StoreAPI struct {
	Mongo  *mongo.Service `inject:""`
	Store  *store.Module `inject:""`
}

func (self *StoreAPI) PlaceOrder(c *gin.Context) {

	var form OrderForm

	if c.BindJSON(&form) == nil {

		order := store.OrderModel{
			User: store.OrderUserModel{
				Name: form.User.Name,
				Email: form.User.Email,
				Phone: form.User.Phone,
			},
			Content: form.Content,
			Budget: form.Budget,
			Currency: form.Currency,
			Games: form.Games,
			Extra: form.Extra,
			BuyDelay: form.BuyDelay,
		}

		self.Store.CreateOrder(order)

		c.JSON(200, gin.H{"status": "okay"})
	}
}

type OrderForm struct {
    User     OrderUserForm `json:"user" binding:"required"`
    Content  string   `json:"content" binding:"required"`
    Budget   int      `json:"budget" binding:"required"`
    Currency string   `json:"currency" binding:"required"`
    BuyDelay int      `json:"buydelay" binding:"required"`
    Games    []string `json:"games"`
    Extra    []string `json:"extra"`
}

type OrderUserForm struct {
	Name string `json:"name" binding:"required"` 
	Email string `json:"email" binding:"required"` 
	Phone string `json:"phone" binding:"required"` 
}