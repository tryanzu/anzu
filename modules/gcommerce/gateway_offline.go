type gcommerce 

type GatewayOffline struct {
	di *Module
}

// Set DI instance
func (this *GatewayOffline) SetDI(di *Module) {
	this.di = di
}

func (this *GatewayOffline) Charge() {
	
}

func (this *GatewayOffline) ModifyPrice(p float64) float64 {
	return p
}