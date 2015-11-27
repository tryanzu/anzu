package cart

type Cart struct {
	items map[string]*CartItem
	storage CartBucket
}

type mutator func(*CartItem) 

func Boot(storage CartBucket) (*Cart, error) {

	restored, err := storage.Restore()

	if err != nil {
		return nil, err
	}

	bucket := &Cart{
		items: restored,
		storage: storage,
	}

	return bucket, nil
}

// A Cart will have many items on it and this will write to it.
func (module *Cart) Add(id, name string, price float64, q int, attrs map[string]interface{}) *CartItem {

	module.items[id] = &CartItem{
		Id: id,
		Name: name,
		Price: price,
		Quantity: q,
		Attributes: attrs,
	}

	module.storage.Save(module.items)

	return module.items[id]
}

// An Item will be removed of the list in case it exists.
func (module *Cart) Remove(id string) bool {

	if _, exists := module.items[id]; exists {

		delete(module.items, id)
		module.storage.Save(module.items)

		return true
	}

	return false
}

// IsEmpty checks if no items in cart object.
func (module *Cart) IsEmpty() bool {

	if len(module.items) > 0 {
		return false
	} else {
		return true
	}
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

	module.storage.Save(module.items)
}