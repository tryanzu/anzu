package payments

type Gateway interface {
	SetOptions(map[string]interface{})
	GetName() string
}

// Gets the related gateway interface to name
func (m *Module) GetGateway(n string) Gateway {

	switch n {
	case "paypal":
		return &Paypal{}
	default:
		panic("Unexpected gateway name.")
		return nil
	}
}
