package controller 

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/gin-gonic/gin"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/gin-gonic/contrib/sessions"
)

type CartAPI struct {
	Components *components.Module `inject:""`
}

// Add Cart item from component id
func (this CartAPI) Add(c *gin.Context) {

	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {

		c.JSON(400, gin.H{"message": "Invalid request, check id format.", "status": "error"})
		return
	}

	session := sessions.Default(c)
	var count int
    v := session.Get("count")
    if v == nil {
      count = 0
    } else {
      count = v.(int)
      count += 1
    }
    session.Set("count", count)
    session.Save()
    c.JSON(200, gin.H{"count": count})
}