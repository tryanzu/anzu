package handle

import (
	"github.com/fernandez14/spartangeek-blacker/modules/gaming"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

type GamingAPI struct {
	Gaming *gaming.Module `inject:""`
}

// Get gamification rules (roles, badges)
func (di *GamingAPI) GetRules(c *gin.Context) {

	rules := di.Gaming.GetRules()

	// Just return the previously loaded rules
	c.JSON(200, rules)
}

// Get users ranking
func (di *GamingAPI) GetRanking(c *gin.Context) {

	sort := c.DefaultQuery("sort", "swords")

	if sort != "swords" && sort != "badges" && sort != "wealth" {

		c.JSON(400, gin.H{"status": "error", "message": "Invalid sort option."})
		return
	}

	ranking := di.Gaming.GetRankingBy(sort)

	c.JSON(200, ranking)
}

func (di *GamingAPI) BuyBadge(c *gin.Context) {

	id := c.Params.ByName("id")
	user_string_id := c.MustGet("user_id")
	user_id := bson.ObjectIdHex(user_string_id.(string))

	// Check badge id
	if bson.IsObjectIdHex(id) == false {

		c.JSON(400, gin.H{"status": "error", "message": "Invalid badge id."})
		return
	}

	// Get the user using gaming module
	user := di.Gaming.Get(user_id)

	// Get the badge we want to adquire
	err := user.AcquireBadge(bson.ObjectIdHex(id), true)

	if err != nil {

		c.JSON(400, gin.H{"status": "error", "message": err})
		return
	}

	c.JSON(200, gin.H{"status": "okay"})
}
