package cart

type CartBucket interface {

	// Restore the cart at runtime.
	Restore() (map[string]*CartItem, error)

	// Save the cart struct for persistance.
	Save(interface{}) error
}
