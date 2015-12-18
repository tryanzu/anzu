package gcommerce

type GatewayOffline struct {
	di    *Module
	order *Order
}

// Set DI instance
func (this *GatewayOffline) SetDI(di *Module) {
	this.di = di
}

func (this *GatewayOffline) SetOrder(order *Order) {
	this.order = order
}

func (this *GatewayOffline) Charge(amount float64) error {
	return nil
}

func (this *GatewayOffline) ModifyPrice(p float64) float64 {
	return p
}
