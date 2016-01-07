package cart

type Cart struct {
	items   map[string]ItemFoundation
	storage CartBucket
}

type mutator func(ItemFoundation)

func Boot(storage CartBucket) (*Cart, error) {

	restored, err := storage.Restore()

	if err != nil {
		return nil, err
	}

	items := map[string]ItemFoundation{}

	for id, item := range restored {

		items[id] = item
	}

	bucket := &Cart{
		items:   items,
		storage: storage,
	}

	return bucket, nil
}

func (module *Cart) Add(item ItemFoundation) {

	id := item.GetId()

	if _, exists := module.items[id]; !exists {

		module.items[id] = item
	} else {

		module.items[id].IncQuantity(1)
	}

	err := module.storage.Save(module.items)

	if err != nil {
		panic(err)
	}
}

// Update an Item inside the cart
func (module *Cart) Update(item ItemFoundation) {

	id := item.GetId()

	module.items[id] = item

	err := module.storage.Save(module.items)

	if err != nil {
		panic(err)
	}
}

// An Item will be removed of the list in case it exists.
func (module *Cart) Remove(id string) bool {

	if _, exists := module.items[id]; exists {

		module.items[id].IncQuantity(-1)

		if module.items[id].GetQuantity() <= 0 {

			delete(module.items, id)
		}

		err := module.storage.Save(module.items)

		if err != nil {
			panic(err)
		}

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
func (module *Cart) GetContent() map[string]ItemFoundation {
	return module.items
}

func (module *Cart) Each(callback mutator) {

	for _, item := range module.items {

		// Execute mutator passed to Each method
		callback(item)
	}

	err := module.storage.Save(module.items)

	if err != nil {
		panic(err)
	}
}
