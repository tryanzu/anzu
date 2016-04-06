package cart

import (
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
)

func (this API) generateSiftItem(p *gcommerce.Product) map[string]interface{} {

	micros := int64((p.Price * 100) * 10000)
	manufacturer, _ := p.Attrs["manufacturer"].(string)

	data := map[string]interface{}{
		"$item_id": p.Id.Hex(),
		"$product_title": p.Name,
		"$price": micros,
		"$currency_code": "MXN",
		"$brand": manufacturer,
		"$manufacturer": manufacturer,
		"$category": p.Type + "-" + p.Category,
		"$quantity": 1,
	}

	return data
}