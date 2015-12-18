package cart

import (
	"errors"
)

type CartItem struct {
	Id         string                 `json:"id"`
	Name       string                 `json:"name"`
	Price      float64                `json:"price"`
	Quantity   int                    `json:"quantity"`
	Attributes map[string]interface{} `json:"attrs"`
}

func (item *CartItem) Attr(name string) (interface{}, error) {

	if attr, exists := item.Attributes[name]; exists {

		return attr, nil
	}

	return nil, errors.New("Attribute does not exists.")
}
