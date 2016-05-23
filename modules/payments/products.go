package payments

type Product interface {
	GetName() string
	GetDescription() string
	GetQuantity() int
	GetCurrency() string
	GetPrice() float64
}

type DigitalProduct struct {
}

type PhysicalProduct struct {
}
