package gcommerce

import (
	"errors"
	"fmt"
	"time"

	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/mgo.v2/bson"
)

// Set DI instance
func (this *Customer) SetDI(di *Module) {
	this.di = di
}

func (this *Customer) GetCart() ([]Cart, error) {

	var items []Cart

	database := this.di.Mongo.Database
	err := database.C("gcommerce_cart_items").Find(bson.M{"customer_id": this.Id, "deleted_at": bson.M{"$exists": false}}).All(&items)

	if err != nil {
		return items, err
	}

	if len(items) == 0 {
		return items, errors.New("No items for this user")
	}

	return items, nil
}

func (this *Customer) GetUser() *user.One {

	if this.User == nil {

		usr, err := this.di.User.Get(this.UserId)

		fmt.Println("User from customer...")
		fmt.Printf("%v\n", this)

		if err != nil {
			panic(err)
		}

		this.User = usr
	}

	return this.User
}

func (this *Customer) MAddresses() {

	var ls []CustomerAddress

	database := this.di.Mongo.Database
	err := database.C("customer_addresses").Find(bson.M{"customer_id": this.Id}).All(&ls)

	if err != nil {
		panic(err)
	}

	// Dump objects inside the same customer pointer
	this.Addresses = ls
}

func (this *Customer) Address(id bson.ObjectId) (*CustomerAddress, error) {

	var a *CustomerAddress

	database := this.di.Mongo.Database
	err := database.C("customer_addresses").Find(bson.M{"customer_id": this.Id, "_id": id}).One(&a)

	if err != nil {
		return a, errors.New("Not found")
	}

	// Share DI instance with address object
	a.SetDI(this.di)

	return a, nil
}

func (this *Customer) AddAddress(name, country, state, city, postal_code, neighborhood, line1, line2, extra, recipient, phone string) CustomerAddress {

	a := Address{
		Country:      country,
		State:        state,
		City:         city,
		PostalCode:   postal_code,
		Line1:        line1,
		Line2:        line2,
		Extra:        extra,
		Neighborhood: neighborhood,
	}

	ca := CustomerAddress{
		Id:         bson.NewObjectId(),
		CustomerId: this.Id,
		Alias:      name,
		Slug:       helpers.StrSlug(name),
		Address:    a,
		TimesUsed:  0,
		LastUsed:   time.Now(),
		Default:    false,
		Recipient:  recipient,
		Phone:      phone,
		Created:    time.Now(),
		Updated:    time.Now(),
	}

	database := this.di.Mongo.Database
	err := database.C("customer_addresses").Insert(ca)

	if err != nil {
		panic(err)
	}

	return ca
}

func (this *Customer) DeleteAddress(id bson.ObjectId) error {

	database := this.di.Mongo.Database
	err := database.C("customer_addresses").RemoveId(id)

	return err
}

func (this *Customer) UpdateAddress(id bson.ObjectId, name, country, state, city, postal_code, neighborhood, line1, line2, extra, recipient, phone string) (CustomerAddress, error) {

	var a CustomerAddress

	database := this.di.Mongo.Database
	err := database.C("customer_addresses").Find(bson.M{"customer_id": this.Id, "_id": id}).One(&a)

	if err != nil {
		return a, errors.New("Not found")
	}

	ad := Address{
		Country:      country,
		State:        state,
		City:         city,
		PostalCode:   postal_code,
		Line1:        line1,
		Line2:        line2,
		Neighborhood: neighborhood,
		Extra:        extra,
	}

	set := bson.M{"address": ad, "updated_at": time.Now(), "alias": name, "slug": helpers.StrSlug(name), "recipient": recipient, "phone": phone}
	err = database.C("customer_addresses").Update(bson.M{"customer_id": this.Id, "_id": id}, bson.M{"$set": set})

	if err != nil {
		panic(err)
	}

	a.Address = ad
	a.Updated = time.Now()
	a.Alias = name
	a.Slug = helpers.StrSlug(name)

	return a, nil
}

func (this *Customer) NewOrder(gateway_name string, meta map[string]interface{}) (*Order, error) {

	t := time.Now()

	reference := t.Format("20060102-1504") + "-" + helpers.StrCapRandom(5)
	order := &Order{
		Id:        bson.NewObjectId(),
		Reference: reference,
		Status:    ORDER_AWAITING,
		Statuses:  make([]Status, 0),
		UserId:    this.Id,
		Items:     make([]Item, 0),
		Total:     0,
		Gateway:   gateway_name,
		Meta:      meta,
		Created:   time.Now(),
		Updated:   time.Now(),
	}

	order.SetDI(this.di)

	return order, nil
}

func (this *Customer) MassdropTransaction(product *Product, q int, gateway_name string, meta map[string]interface{}) (*Order, *MassdropTransaction, error) {

	if product.Massdrop == nil {
		return nil, nil, errors.New("Can't make a massdrop transaction using this product.")
	}

	meta["skip_siftscience"] = true
	meta["addressless"] = true

	t := time.Now()

	reference := t.Format("20060102-1504") + "-MASSD-" + helpers.StrCapRandom(5)
	order := &Order{
		Id:        bson.NewObjectId(),
		Reference: reference,
		Status:    ORDER_AWAITING,
		Statuses:  make([]Status, 0),
		UserId:    this.Id,
		Items:     make([]Item, 0),
		Total:     0,
		Gateway:   gateway_name,
		Meta:      meta,
		Created:   time.Now(),
		Updated:   time.Now(),
	}

	order.SetDI(this.di)

	// Add reservation as an item
	item_name := "Reservación de producto en legión (" + product.Name + ")"
	item_meta := map[string]interface{}{
		"related":    "massdrop_product",
		"related_id": product.Id,
	}

	order.Add(item_name, product.Description, product.Image, product.Massdrop.Reserve, q, item_meta)

	err := order.Save()

	if err != nil {
		return nil, nil, err
	}

	err = order.Checkout()

	if err != nil {
		return nil, nil, err
	}

	var transactionType string = "reservation"

	if gateway_name == "offline" {
		transactionType = MASSDROP_TRANS_INSTERESTED
	}

	transaction := &MassdropTransaction{
		Id:         bson.NewObjectId(),
		MassdropId: product.Massdrop.Id,
		CustomerId: this.Id,
		Type:       transactionType,
		Status:     MASSDROP_STATUS_COMPLETED,
		Attrs: map[string]interface{}{
			"order_id": order.Id,
			"quantity": q,
		},
		Created: time.Now(),
		Updated: time.Now(),
	}

	transaction.SetDI(this.di)
	transaction.Save()

	return order, transaction, nil
}
