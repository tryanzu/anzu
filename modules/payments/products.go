package payments

type Product interface {
	GetName() string
	GetDescription() string
	GetQuantity() int
	GetCurrency() string
	GetPrice() float64
}

type DigitalProduct struct {
	Name        string
	Description string
	Quantity    int
	Currency    string
	Price       float64
}

func (d *DigitalProduct) GetName() string {
	return d.Name
}

func (d *DigitalProduct) GetDescription() string {
	return d.Description
}

func (d *DigitalProduct) GetQuantity() int {
	return d.Quantity
}

func (d *DigitalProduct) GetCurrency() string {
	return d.Currency
}

func (d *DigitalProduct) GetPrice() float64 {
	return d.Price
}

type PhysicalProduct struct {
}
