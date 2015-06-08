package handle

import (
	"github.com/gin-gonic/gin"
	//"github.com/gin-gonic/gin/binding"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"github.com/fernandez14/spartangeek-blacker/model"
	"gopkg.in/mgo.v2/bson"
)

type PartAPI struct {
	Data *mongo.Service `inject:""`
}

func (di *PartAPI) GetPartTypes(c *gin.Context) {

	// Get the database service from the DI container
	database  := di.Data.Database

	var types []string

	err := database.C("parts").Find(nil).Distinct("type", &types)
	if err != nil {
		panic(err)
	}

	c.JSON(200, gin.H{"status": "okay", "types": types})
}

func (di *PartAPI) GetPartManufacturerModels(c *gin.Context) {

	// Get the database service from the DI container
	database  := di.Data.Database
	part_type := c.Param("type")
	manufacturer := c.Query("manufacturer")

	var parts []model.PartByModel

	err := database.C("parts").Find(bson.M{"type": part_type, "manufacturer": manufacturer}).Select(bson.M{"_id": 1, "name": 1}).All(&parts)
	if err != nil {
		panic(err)
	}

	c.JSON(200, gin.H{"status": "okay", "parts": parts})
}

func (di *PartAPI) GetPartManufacturers(c *gin.Context) {

	// Get the database service from the DI container
	database  := di.Data.Database
	part_type := c.Param("type")

	var manufacturers []string

	err := database.C("parts").Find(bson.M{"type": part_type}).Distinct("manufacturer", &manufacturers)
	if err != nil {
		panic(err)
	}

	c.JSON(200, gin.H{"status": "okay", "manufacturers": manufacturers})
}