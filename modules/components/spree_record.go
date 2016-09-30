package components

import (
	"gopkg.in/mgo.v2/bson"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SpreeRecord struct {
	Id          bson.ObjectId `bson:"_id,omitempty" json:"id"`
	SpreeId     int           `bson:"spree_id"`
	ComponentId bson.ObjectId `bson:"component_id"`
	Type        string        `bson:"type"`
	Images      []string      `bson:"images"`
	Properties  []string      `bson:"properties"`
	Created     time.Time     `bson:"created_at"`
	Updated     time.Time     `bson:"updated_at"`

	di *Module
}

type SpreeStocks struct {
	Count int                      `json:"count"`
	Items []map[string]interface{} `json:"stock_items"`
}

var taxons map[string]string = map[string]string{
	"video-card": "2",
	"cpu":        "3",
	"storage":    "4",
	"monitor":    "6",
	"case":       "9",
	"keyboard":   "5",
	"mouse":      "5",
}

func (spree *SpreeRecord) SetDI(di *Module) {
	spree.di = di
}

func (spree *SpreeRecord) UpdatePrice(p float64) {

	form := url.Values{}
	form.Add("product[price]", strconv.FormatFloat(p, 'f', 0, 64))
	req, _ := http.NewRequest("PUT", SPREE_API_URL+"products/"+strconv.Itoa(spree.SpreeId), strings.NewReader(form.Encode()))
	client := http.Client{}

	req.Header.Add("X-Spree-Token", SPREE_TOKEN)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("Could not update price (%v): %v\n\nAborting....\n", resp.StatusCode, string(body)))
	}
}

func (spree *SpreeRecord) UpdateStock(status bool) {

	if status == true {
		spree.ensureStockItems()
	} else {
		spree.ensureNoStock()
	}
}

func (spree *SpreeRecord) UpdateTaxons() {

	if taxon, exists := taxons[spree.Type]; exists {

		form := url.Values{}
		form.Add("product[taxon_ids][]", taxon)

		req, _ := http.NewRequest("PUT", SPREE_API_URL+"products/"+strconv.Itoa(spree.SpreeId), strings.NewReader(form.Encode()))
		client := http.Client{}

		req.Header.Add("X-Spree-Token", SPREE_TOKEN)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)

		if err != nil {
			panic(err)
		}

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			panic(err)
		}

		if resp.StatusCode != 200 {
			panic(fmt.Sprintf("Could not post product (%v): %v\n\nAborting....\n", resp.StatusCode, string(body)))
		}
	}
}

func (spree *SpreeRecord) fetchStockItems() SpreeStocks {

	req, _ := http.NewRequest("GET", SPREE_API_URL+"stock_locations/1/stock_items?q[variant_id_eq]="+strconv.Itoa(spree.SpreeId), nil)
	req.Header.Add("X-Spree-Token", SPREE_TOKEN)

	client := http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic("Could not fetch stock items")
	}

	var response SpreeStocks
	err = json.Unmarshal(body, &response)

	if err != nil {
		panic(err)
	}

	return response
}

func (spree *SpreeRecord) ensureStockItems() {
	stock := spree.fetchStockItems()
	spree.di.Logger.Debugf("%+v\n", stock)

	if stock.Count == 0 {

		payload := []byte(`{
		  "stock_item": {
		    "count_on_hand": "10",
		    "variant_id": "` + strconv.Itoa(spree.SpreeId) + `",
		    "backorderable": true
		  }
		}`)

		req, _ := http.NewRequest("POST", SPREE_API_URL+"stock_locations/1/stock_items", bytes.NewBuffer(payload))
		req.Header.Add("X-Spree-Token", SPREE_TOKEN)
		req.Header.Set("Content-Type", "application/json")

		client := http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			panic(err)
		}

		body, _ := ioutil.ReadAll(resp.Body)

		spree.di.Logger.Debugf("%+v\n", string(body))

		if resp.StatusCode != 201 {
			panic("Did not receive valid status code while creating stocks")
		}
	}
}

func (spree *SpreeRecord) ensureNoStock() {
	stock := spree.fetchStockItems()
	spree.di.Logger.Debugf("%+v\n", stock)

	if stock.Count > 0 && len(stock.Items) > 0 {
		id, exists := stock.Items[0]["id"]

		if exists {
			req, _ := http.NewRequest("DELETE", SPREE_API_URL+"stock_locations/1/stock_items/"+fmt.Sprintf("%.0f", id.(float64)), nil)
			req.Header.Add("X-Spree-Token", SPREE_TOKEN)

			client := http.Client{}
			resp, err := client.Do(req)

			if err != nil {
				panic(err)
			}

			body, _ := ioutil.ReadAll(resp.Body)

			spree.di.Logger.Debugf("%+v\n", string(body))

			if resp.StatusCode != 204 {
				panic("Did not receive valid status code while deleting stocks")
			}
		}
	}
}
