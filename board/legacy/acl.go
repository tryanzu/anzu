package handle

import (
	"github.com/fernandez14/spartangeek-blacker/modules/acl"
	"github.com/gin-gonic/gin"
)

type AclAPI struct {
	Acl *acl.Module `inject:""`
}

func (di *AclAPI) GetRules(c *gin.Context) {

	rules := di.Acl.Rules

	c.JSON(200, gin.H{"status": "okay", "rules": rules})
}
