package gcommerce

// Set DI instance
func (this *Order) SetDI(di *Module) {
	this.di = di
}

func (this *Order) Add(name, description, image string, price float64, q int, meta map[string]interface{}) {

	this.Items = append(this.Items, Item{name, image, description, price, q, meta})
}