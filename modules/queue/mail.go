package queue 

import (
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/mail"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/mgo.v2/bson"
	"github.com/iron-io/iron_go3/mq"
	
	"log"
)

type MailJob struct {
	GCommerce  *gcommerce.Module  `inject:""`
	Mail  *mail.Module   `inject:""`
	User  *user.Module   `inject:""`
}

func (self MailJob) StoreDelayedResponse(params map[string]interface{}) {
		
	log.Println("Done from StoreDelayedResponse")
}

func (self MailJob) GcommercePayReminder(params map[string]interface{}) {
	
	order_id, exists := params["id"]
	
	if exists {
		
		if bson.IsObjectIdHex(order_id.(string)) == false {
			log.Println("[error] Invalid GcommercePayReminder job")
			return
		}
	
		id := bson.ObjectIdHex(order_id.(string))
		order, err := self.GCommerce.One(bson.M{"_id": id})
		
		if err != nil {
			log.Println("[error] Invalid GcommercePayReminder referenced order: not found")
			return	
		}
		
		if order.Status != gcommerce.ORDER_AWAITING {
			
			// The order has been confirmed or something
			return
		}
		
		customer, err := self.GCommerce.GetCustomer(order.UserId)

		if err != nil {
			log.Println("[error] Invalid GcommercePayReminder referenced order: customer not found")
			return
		}
	
		usr, err := self.User.Get(customer.UserId)
	
		if err != nil {
			log.Println("[error] Invalid GcommercePayReminder referenced order: usr not found")
			return
		}
		
		address_id, exists := order.Shipping.Meta["related_id"]
		
		if exists {
			
			address, err := customer.Address(address_id.(bson.ObjectId))
			
			if err != nil {
				log.Println("[error] Invalid GcommercePayReminder referenced order: address not found")
				return
			}
			
			var paymentType string

			if order.Gateway == "offline" {
				paymentType = "Transferencia o Deposito"
			} else if order.Gateway == "stripe" {
				paymentType = "Pago en linea"
			}
		
			mailing := self.Mail 
			{
				compose := mail.Mail{
					Template: 361601,
					FromName: "Spartan Geek",
					FromEmail: "pedidos@spartangeek.com",
					Recipient: []mail.MailRecipient{
						{
							Name:  usr.Name(),
							Email: usr.Email(),
						},
						{
							Name:  "Equipo Spartan Geek",
							Email: "pedidos@spartangeek.com",
						},
					},
					Variables: map[string]interface{}{
						"name": usr.Name(),
						"payment": paymentType,
						"line1": address.Line1(),
						"line2": address.Line2(),
						"line3": address.Extra(),
						"total_products": order.GetOriginalTotal() - order.Shipping.OPrice,
						"total_shipping": order.Shipping.OPrice,
						"subtotal": order.GetOriginalTotal(),
						"commision": order.GetGatewayCommision(),
						"total": order.Total,
						"items": order.Items,
						"reference": order.Reference,
					},
				}
				
				mailing.Send(compose)
				
				// Replicate the mail for two days
				q := mq.New("gcommerce")
				_, err := q.PushMessage(mq.Message{Delay: 3600*24*2, Body: "{\"fire\":\"payment-reminder\",\"id\":\""+order.Id.Hex()+"\"}"})
				
				if err != nil {
					log.Printf("[error][check] Couldnt replicate GcommercePayReminder for order: %v\n", order.Id.Hex())
				}
			}
		}
	}
}