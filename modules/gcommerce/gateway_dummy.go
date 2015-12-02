type gcommerce 

type GatewayDummy struct {
	di *Module
}

// Set DI instance
func (this *GatewayDummy) SetDI(di *Module) {
	this.di = di
}

func (this *GatewayDummy) Charge() {
	
}

func (this *GatewayDummy) ModifyPrice(p float64) float64 {
	return p
}