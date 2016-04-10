package gcommerce

import (
	"gopkg.in/mgo.v2/bson"

	"time"
)

// Set DI instance
func (this *MassdropTransaction) SetDI(di *Module) {
	this.di = di
}

func (this *MassdropTransaction) Save() error {

	database := this.di.Mongo.Database

	// Perform the save of the order once we've got here
	err := database.C("gcommerce_massdrop_transactions").Insert(this)

	return err
}

func (this *Product) MassdropInterested(user_id bson.ObjectId) bool {

	if this.Massdrop == nil {
		return false
	}

	var model MassdropTransaction

	database := this.di.Mongo.Database
	customer := this.di.GetCustomerFromUser(user_id)
	err := database.C("gcommerce_massdrop_transactions").Find(bson.M{"massdrop_id": this.Massdrop.Id, "customer_id": customer.Id, "status": MASSDROP_STATUS_COMPLETED}).One(&model)

	if err == nil {

		err := database.C("gcommerce_massdrop_transactions").Update(bson.M{"_id": model.Id}, bson.M{"$set": bson.M{"status": MASSDROP_STATUS_REMOVED, "updated_at": time.Now()}})

		if err != nil {
			panic(err)
		}

		return false

	} else {

		transaction := &MassdropTransaction{
			Id:          bson.NewObjectId(),
			MassdropId:  this.Massdrop.Id,
			CustomerId:  customer.Id,
			Type:        MASSDROP_TRANS_INSTERESTED,         
			Status:      MASSDROP_STATUS_COMPLETED,
			Attrs:       map[string]interface{}{},
			Created:     time.Now(),
			Updated:     time.Now(),
		}

		transaction.SetDI(this.di)
		transaction.Save()

		return true
	}
}

func (this *Product) UserMassdrop(user_id bson.ObjectId) {

	if this.Massdrop == nil {
		return
	}

	var model MassdropTransaction
	var status string

	database := this.di.Mongo.Database
	customer := this.di.GetCustomerFromUser(user_id)
	err := database.C("gcommerce_massdrop_transactions").Find(bson.M{"massdrop_id": this.Massdrop.Id, "customer_id": customer.Id, "status": MASSDROP_STATUS_COMPLETED}).One(&model)

	if err == nil {
		status = model.Type
	} else {
		status = "none"
	}

	this.Massdrop.Current = status
}