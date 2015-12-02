package gcommerce 

type GatewayStripe struct {
	di *Module
}

// Set DI instance
func (this *GatewayStripe) SetDI(di *Module) {
	this.di = di
}

func (this *GatewayStripe) Charge() {
	
}

func (this *GatewayStripe) ModifyPrice(p float64) float64 {
	return p * 1.035
}