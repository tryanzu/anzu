package gcommerce

import (
	"github.com/fernandez14/spartangeek-blacker/modules/user"
	"gopkg.in/mgo.v2/bson"

	"errors"
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

func (this *MassdropTransaction) CastToReservation() error {

	database := this.di.Mongo.Database

	// Perform the save of the order once we've got here
	err := database.C("gcommerce_massdrop_transactions").Update(bson.M{"_id": this.Id}, bson.M{"$set": bson.M{"status": MASSDROP_STATUS_COMPLETED, "type": MASSDROP_TRANS_RESERVATION}})

	return err
}

func (this *Product) MassdropInterested(user_id bson.ObjectId, reference string) (bool, error) {

	if this.Massdrop == nil {
		return false, nil
	}

	var model *MassdropTransaction

	database := this.di.Mongo.Database
	customer := this.di.GetCustomerFromUser(user_id)
	err := database.C("gcommerce_massdrop_transactions").Find(bson.M{"massdrop_id": this.Massdrop.Id, "customer_id": customer.Id, "status": MASSDROP_STATUS_COMPLETED}).One(&model)

	if err == nil {

		err := database.C("gcommerce_massdrop_transactions").Update(bson.M{"_id": model.Id}, bson.M{"$set": bson.M{"status": MASSDROP_STATUS_REMOVED, "updated_at": time.Now()}})

		if err != nil {
			return false, err
		}

		return false, nil

	} else {

		if len(reference) == 0 {

			return false, errors.New("Invalid reference, must have one")
		}

		transaction := &MassdropTransaction{
			Id:         bson.NewObjectId(),
			MassdropId: this.Massdrop.Id,
			CustomerId: customer.Id,
			Type:       MASSDROP_TRANS_INSTERESTED,
			Status:     MASSDROP_STATUS_COMPLETED,
			Attrs: map[string]interface{}{
				"reference": reference,
			},
			Created: time.Now(),
			Updated: time.Now(),
		}

		transaction.SetDI(this.di)
		transaction.Save()

		return true, nil
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
	acl := this.di.Acl.User(user_id)
	err := database.C("gcommerce_massdrop_transactions").Find(bson.M{"massdrop_id": this.Massdrop.Id, "customer_id": customer.Id, "status": MASSDROP_STATUS_COMPLETED}).Sort("-created_at").One(&model)

	if err == nil {
		status = model.Type
	} else {
		status = "none"
	}

	if acl.Can("sensitive-data") && this.Massdrop.usersList != nil {

		umap := this.Massdrop.usersList.(map[bson.ObjectId]user.UserBasic)
		list := []user.UserBasic{}

		for _, u := range umap {
			list = append(list, u)
		}

		this.Massdrop.Users = list
	}

	this.Massdrop.Current = status
}
