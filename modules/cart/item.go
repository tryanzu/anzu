package cart

import (
	"errors"
)

type ItemFoundation interface {
	GetId() string
	GetName() string
	GetPrice() float64
	SetPrice(float64)
	GetQuantity() int
	IncQuantity(int)
	Attr(string) (interface{}, error)
}

type CartItem struct {
	Id          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Image       string                 `json:"image"`
	Price       float64                `json:"price"`
	Quantity    int                    `json:"quantity"`
	Type        string				   `json:"type"`
	Attributes  map[string]interface{} `json:"attrs"`
}

func (item *CartItem) GetId() string {
	return item.Id
}

func (item *CartItem) GetName() string {
	return item.Name
}

func (item *CartItem) GetPrice() float64 {
	return item.Price
}

func (item *CartItem) SetPrice(p float64) {
	item.Price = p
}

func (item *CartItem) GetQuantity() int {
	return item.Quantity
}

func (item *CartItem) IncQuantity(by int) {
	item.Quantity = item.Quantity + by
}

func (item *CartItem) Attr(name string) (interface{}, error) {

	if attr, exists := item.Attributes[name]; exists {

		return attr, nil
	}

	return nil, errors.New("Attribute does not exists.")
}
