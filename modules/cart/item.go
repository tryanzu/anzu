package cart

type CartItem struct {
	Id string
	Name string
	Price float 
	Quantity int
	Attributes map[string]interface{}
}

