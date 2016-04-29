package cli

import (
	"github.com/fernandez14/go-enlacefiscal"
	"github.com/fernandez14/spartangeek-blacker/modules/gcommerce"
	"github.com/fernandez14/spartangeek-blacker/modules/store"
	"gopkg.in/mgo.v2/bson"

	"fmt"
	"io/ioutil"
	"strconv"
	"time"
)

func (module Module) GenerateMassdropInvoices() {

	var transactions []gcommerce.MassdropTransaction
	var orders_id []bson.ObjectId
	var orders []gcommerce.Order
	var invoices map[string][]gcommerce.Order

	database := module.Mongo.Database
	err := database.C("gcommerce_massdrop_transactions").Find(bson.M{"type": gcommerce.MASSDROP_TRANS_RESERVATION, "status": gcommerce.MASSDROP_STATUS_COMPLETED}).All(&transactions)

	if err != nil {
		panic(err)
	}

	fmt.Println("Found " + strconv.Itoa(len(transactions)) + " massdrop candidate transactions.")

	for _, t := range transactions {

		if order_id, exists := t.Attrs["order_id"].(bson.ObjectId); exists {
			orders_id = append(orders_id, order_id)
		}
	}

	err = database.C("gcommerce_orders").Find(bson.M{"status": "confirmed", "_id": bson.M{"$in": orders_id}}).All(&orders)

	if err != nil {
		panic(err)
	}

	fmt.Println("Found " + strconv.Itoa(len(orders)) + " related massdrop orders that are invoice eligible")
	fmt.Println("Enter the massdrop settlement order related product:")

	var productId string
	fmt.Scanln(&productId)

	if !bson.IsObjectIdHex(productId) {
		fmt.Println("The entered product id is not valid, it'll skip the invoice joining...")
	}

	invoices = make(map[string][]gcommerce.Order, 0)

	for _, o := range orders {

		invoices[o.UserId.Hex()] = []gcommerce.Order{o}

		if bson.IsObjectIdHex(productId) {

			var settlement gcommerce.Order

			err := database.C("gcommerce_orders").Find(bson.M{"customer_id": o.UserId, "status": "confirmed", "items.meta.related_id": bson.ObjectIdHex(productId)}).One(&settlement)

			if err == nil {
				invoices[o.UserId.Hex()] = append(invoices[o.UserId.Hex()], settlement)
			}
		}
	}

	fmt.Println("Have found " + strconv.Itoa(len(invoices)) + " single invoices to emit")

	for user, list := range invoices {

		fmt.Println(user + " has " + strconv.Itoa(len(list)) + " invoices")
	}
}

func (module Module) GenerateCustomInvoice() {

	config, err := module.Config.Get("invoicing")

	if err != nil {
		panic(err)
	}

	apiUser, err := config.String("username")

	if err != nil {
		panic(err)
	}

	apiPass, err := config.String("password")

	if err != nil {
		panic(err)
	}

	rfcOrigin, err := config.String("rfc")

	if err != nil {
		panic(err)
	}

	series, err := config.String("series")

	if err != nil {
		panic(err)
	}

	folioPath, err := config.String("folio")

	if err != nil {
		panic(err)
	}

	folioContent, err := ioutil.ReadFile(folioPath)

	if err != nil {
		panic(err)
	}

	folio, err := strconv.Atoi(string(folioContent))

	if err != nil {
		panic(err)
	}

	api := efiscal.Boot(apiUser, apiPass, false)
	invoice := api.Invoice(rfcOrigin, series, strconv.Itoa(folio))

	receiver := &efiscal.Receiver{
		"VACS840206QC7",
		"SILVERIO VALDEZ CUEVAS",
		&efiscal.Address{
			Street:       "ALFONSO SIERRA PARTIDA",
			Ext:          "#192",
			Int:          "-",
			Neighborhood: "COLONIA JARDINES DE VISTA HERMOSA",
			Locality:     "Colima",
			Town:         "Colima",
			State:        "Colima",
			Country:      "México",
			Zipcode:      "28017",
		},
	}
	invoice.AddItem(efiscal.Item{
		Quantity:    1,
		Value:       1637.93,
		Unit:        "producto",
		Description: "Disco duro de 2TB",
	})
	invoice.TransferIVA(16)
	invoice.SetPayment(&efiscal.PAY_ONE_TIME_TRANSFER)
	invoice.SetReceiver(receiver)
	invoice.SendMail([]string{"facturas_comparateca@outlook.com"})

	data, err := api.Sign(invoice)

	var record *store.Invoice

	if err == nil {

		database := module.Mongo.Database
		record = &store.Invoice{
			Id: bson.NewObjectId(),
			Assets: store.InvoiceAssets{
				PDF: "",
				XML: "",
			},
			Meta:    data,
			Created: time.Now(),
			Updated: time.Now(),
		}

		err := database.C("invoices").Insert(record)

		if err != nil {
			panic(err)
		}

		newFolio := strconv.Itoa(folio + 1)
		err = ioutil.WriteFile(folioPath, []byte(newFolio), 0644)

		if err != nil {
			panic(err)
		}
	}
}
