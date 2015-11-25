package cart

type Cart struct {
	items map[string]*CartItem
}

type mutator func(*CartItem) 

// A Cart will have many items on it and this will write to it.
func (module *Cart) Add(id, name string, price float, q int, attrs map[string]interface{}) *CartItem {

	module.items[id] = &CartItem{
		Id: id,
		Name: name,
		Price: price,
		Quantity: q,
		Attributes: attrs,
	}

	return module.items[id]
}

// An Item will be removed of the list in case it exists.
func (module *Cart) Remove(id string) bool {

	if _, exists := module.items[id]; exists {

		delete(module.items, id)
		return true
	}

	return false
}

// IsEmpty checks if no items in cart object.
func (module *Cart) IsEmpty() bool {

	return len(module.items) > 0 ? false : true
}

// Get Cart contents for mutator approaches.
func (module *Cart) GetContent() map[string]*CartItem {

	return module.items
}

func (module *Cart) Each(callback mutator) {

	for _, item := range module.items {

		// Execute mutator passed to Each method
		callback(item)
	}
}