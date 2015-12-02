package gcommerce

// Set DI instance
func (this *Order) SetDI(di *Module) {

	this.di = di

	// Setup gateway
	gateway, err := getGateway(this.Gateway)

	if err != nil {
		panic(err)
	}

	this.gateway = gateway
	this.gateway.SetDI(this.di)
}

func (this *Order) Add(name, description, image string, price float64, q int, meta map[string]interface{}) {

	real_price := price

	this.Items = append(this.Items, Item{name, image, description, real_price, q, meta})
}