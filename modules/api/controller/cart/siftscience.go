package cart

func (this API) generateSiftItem(c CartComponentItem, component *components.ComponentModel) map[string]interface{} {

	micros := int64((c.Price * 100) * 10000)

	data := map[string]interface{}{
		"$item_id": c.Id,
		"$product_title": c.FullName,
		"$price": micros,
		"$currency_code": "MXN",
		"$brand": component.Manufacturer,
		"$manufacturer": component.Manufacturer,
		"$category": component.Type,
		"$quantity": c.Quantity,
	}

	return data
}