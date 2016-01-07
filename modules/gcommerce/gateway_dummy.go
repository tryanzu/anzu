package gcommerce

type GatewayDummy struct {
	di    *Module
	order *Order
	meta  map[string]interface{}
}

// Set DI instance
func (this *GatewayDummy) SetDI(di *Module) {
	this.di = di
}

func (this *GatewayDummy) SetOrder(order *Order) {
	this.order = order
}

func (this *GatewayDummy) SetMeta(meta map[string]interface{}) {
	this.meta = meta
}

func (this *GatewayDummy) Charge(amount float64) error {
	return nil
}

func (this *GatewayDummy) ModifyPrice(p float64) float64 {
	return p
}

func (this *GatewayDummy) AdjustPrice(p float64) float64 {
	return p
}
