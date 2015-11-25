package cart

type CartBucket interface {

	// Restore the cart at runtime.
	Restore() error

	// Save the cart struct for persistance.
	Save() error
}

