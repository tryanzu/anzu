package controller

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/gin-gonic/gin"
)

type CustomerAPI struct {
	Store      *store.Module `inject:""`
	Components *components.Module `inject:""` 
	GCommerce  *gcommerce.Module `inject:""`
}

func (this CustomerAPI) Get(c *gin.Context) {

	user := c.MustGet("user_id")
	userId := bson.ObjectIdHex(user.(string))

	// Load customer
	customer := this.GCommerce.GetCustomerFromUser(userId)

	// Load addresses in customer memory
	customer.MAddresses()

	c.JSON(200, customer)
}

func (this CustomerAPI) CreateAddress(c *gin.Context) {
	
	var form AddressForm
	user := c.MustGet("user_id")
	userId := bson.ObjectIdHex(user.(string))

	if c.Bind(&form) == nil {

		customer := this.GCommerce.GetCustomerFromUser(userId)
		address := customer.AddAddress(form.Alias, "mx", form.State, form.City, form.Zipcode, form.AddressLine1, form.AddressLine2, form.Extra)

		c.JSON(200, gin.H{"status": "okay", "address_id": address.Id})
		return
	}

	c.JSON(400, gin.H{"message": "Invalid request, check customer docs.", "status": "error"})
}

func (this CustomerAPI) DeleteAddress(c *gin.Context) {
	
	address_str := c.Param("id")
	user := c.MustGet("user_id")
	userId := bson.ObjectIdHex(user.(string))

	if !bson.IsObjectIdHex(address_str) {
		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	address_id :=  bson.ObjectIdHex(address_str)

	// Retrieve customer from user
	customer := this.GCommerce.GetCustomerFromUser(userId)

	// Load addresses in customer memory
	customer.MAddresses()

	exists := false

	for _, a := range customer.Addresses {

		if a.Id == address_id {
			exists = true
		}
	}

	if !exists {

		c.JSON(400, gin.H{"message": "Not allowed.", "status": "error"})
		return
	}

	customer.DeleteAddress(address_id)

	c.JSON(200, gin.H{"status": "okay"})
}

func (this CustomerAPI) UpdateAddress(c *gin.Context) {
	
	var form AddressForm
	user := c.MustGet("user_id")
	userId := bson.ObjectIdHex(user.(string))
	address_str := c.Param("id")

	if !bson.IsObjectIdHex(address_str) {
		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	address_id :=  bson.ObjectIdHex(address_str)

	if c.Bind(&form) == nil {

		customer := this.GCommerce.GetCustomerFromUser(userId)
		_, err := customer.UpdateAddress(address_id, form.Alias, "mx", form.State, form.City, form.Zipcode, form.AddressLine1, form.AddressLine2, form.Extra)

		if err != nil {
			c.JSON(400, gin.H{"message": "Couldnt perform operation, try again later.", "status": "error"})
			return
		}

		c.JSON(200, gin.H{"status": "okay"})
	}

	c.JSON(400, gin.H{"message": "Invalid request, check customer docs.", "status": "error"})
}

type AddressForm struct {
	Alias   string `json:"name" binding:"required"`
	State   string `json:"state" binding:"required"`
	City    string `json:"city" binding:"required"`
	Zipcode string `json:"postal_code" binding:"required"`
	AddressLine1 string `json:"line1" binding:"required"`
	AddressLine2 string `json:"line2"`
	Extra        string `json:"extra"`
}