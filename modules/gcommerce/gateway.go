package gcommerce

import (
	"errors"
)

type Gateway interface {

	// Charge money logic
	Charge(float64) error

	// Price mutator
	ModifyPrice(float64) float64

	// Global price mutator
	AdjustPrice(float64) float64

	// Set the Dependency injection module
	SetDI(*Module)
	SetOrder(*Order)
}

func getGateway(name string) (Gateway, error) {

	switch name {

	case "offline":

		return Gateway(&GatewayOffline{}), nil
	case "stripe":

		return Gateway(&GatewayStripe{}), nil
	default:

		return Gateway(&GatewayDummy{}), errors.New("Invalid gateway id.")
	}
}
