package cart

type Cart struct {
	storage CartBucket
}

func Boot(storage CartBucket) (*Cart, error) {

	bucket := &Cart{
		storage: storage,
	}

	return bucket, nil
}

func (cart Cart) Bind(where interface{}) error {
	return cart.storage.Restore(&where)
}

func (cart Cart) Save(data interface{}) error {
	return cart.storage.Save(data)
}