package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/tryanzu/core/modules/acl"
)

type AclAPI struct {
	Acl *acl.Module `inject:""`
}

func (di *AclAPI) GetRules(c *gin.Context) {

	rules := di.Acl.Rules

	c.JSON(200, gin.H{"status": "okay", "rules": rules})
}
