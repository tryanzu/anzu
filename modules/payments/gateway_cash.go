package payments

type Cash struct {
	GenerateInvoice bool
}

func (p *Cash) GetName() string {
	return "cash"
}

func (p *Cash) SetOptions(o map[string]interface{}) {

	if i, exists := o["invoice"]; exists {
		p.GenerateInvoice = i.(bool)
	}
}

func (p *Cash) Purchase(pay *Payment, c *Create) (map[string]interface{}, error) {

	response := map[string]interface{}{}
	pay.Status = PAYMENT_SUCCESS

	return response, nil
}

func (p *Cash) CompletePurchase(pay *Payment, data map[string]interface{}) (map[string]interface{}, error) {

	response := map[string]interface{}{}

	return response, nil
}
