package builds

import (
	"github.com/gin-gonic/gin"
	//"gopkg.in/mgo.v2/bson"
)

func (a API) UpdateAction(c *gin.Context) {

	var m map[string]interface{}

	buildId := c.Param("id")

	if buildId == "" {
		c.JSON(400, gin.H{"status": "error", "message": "Invalid build reference."})
		return
	}

	build, err := a.Builds.FindByRef(buildId)

	if err != nil {
		c.JSON(404, gin.H{"status": "error", "message": "Could not get build by ref."})
		return
	}

	if c.Bind(&m) == nil {
		err := build.UpdateByMap(m)

		if err != nil {
			c.JSON(400, gin.H{"status": "error", "message": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "okay"})
	}
}
