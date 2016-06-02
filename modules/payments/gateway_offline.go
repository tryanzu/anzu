package payments

type Offline struct {
	GenerateInvoice bool
}

func (o *Offline) GetName() string {
	return "offline"
}

func (o *Offline) SetOptions(opts map[string]interface{}) {

	if i, exists := opts["invoice"]; exists {
		o.GenerateInvoice = i.(bool)
	}
}

func (o *Offline) Purchase(pay *Payment, c *Create) (map[string]interface{}, error) {

	response := map[string]interface{}{}
	pay.Status = PAYMENT_AWAITING

	return response, nil
}

func (o *Offline) CompletePurchase(pay *Payment, data map[string]interface{}) (map[string]interface{}, error) {

	response := map[string]interface{}{}

	return response, nil
}
