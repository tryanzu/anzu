package payments

type Gateway interface {

	// Set up gateway options in a generic way
	SetOptions(map[string]interface{})

	// Gets the name of the gateway
	GetName() string

	// Authorize and immediately capture an amount on the customer's card
	Purchase(*Payment, *Create) (map[string]interface{}, error)

	// Handle return from off-site gateways after purchase
	CompletePurchase(*Payment, map[string]interface{}) (map[string]interface{}, error)
}

// Gets the related gateway interface to name
func (m *Module) GetGateway(n string) Gateway {

	if g, e := m.Gateways[n]; e {
		return g
	} else {
		panic("Gateway " + n + " has not been initialized.")
		return nil
	}
}
