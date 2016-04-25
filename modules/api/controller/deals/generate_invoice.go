package deals

import (
	"github.com/gin-gonic/gin"
)

func (this API) GenerateInvoice(c *gin.Context) {

	var form InvoiceForm

	if c.Bind(&form) == nil {

		order_id := form.Id
		order, err := self.Store.Order(order_id)

		if err != nil {
			c.JSON(404, gin.H{"status": "error", "message": "Invalid id, deal not found"})
			return
		}

		invoice, err := order.EmitInvoice(form.Name, form.RFC, form.Total)

	}

	c.JSON(400, gin.H{"status": "error", "message": "Invalid generateInvoice request. Check payload."})
}
