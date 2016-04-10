package gcommerce

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