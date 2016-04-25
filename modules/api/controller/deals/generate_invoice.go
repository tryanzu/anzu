package deals

import (
	"github.com/gin-gonic/gin"
)

func (this API) GenerateInvoice(c *gin.Context) {

	var form InvoiceForm

	if c.Bind(&form) == nil {

	}

	c.JSON(400, gin.H{"status": "error", "message": "Invalid generateInvoice request. Check payload."})
}
