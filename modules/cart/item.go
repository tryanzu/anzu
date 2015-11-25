package cart

type CartItem struct {
	Id string  `json:"id"`
	Name string `json:"name"`
	Price float64 `json:"price"`
	Quantity int `json:"quantity"`
	Attributes map[string]interface{} `json:"attrs"`
}

