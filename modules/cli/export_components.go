package cli

import (
	"gopkg.in/jmcvetta/neoism.v1"
	"gopkg.in/mgo.v2/bson"

	"bytes"
	"fmt"
	"regexp"
)

var camelingRegex = regexp.MustCompile("[0-9A-Za-z]+")

func CamelCase(src string) string {
	byteSrc := []byte(src)
	chunks := camelingRegex.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		chunks[idx] = bytes.Title(val)
	}
	return string(bytes.Join(chunks, nil))
}

func (module Module) ExportComponents() {

	var c map[string]interface{}

	database := module.Mongo.Database
	neo := module.Neoism
	list := database.C("components").Find(nil).Iter()
	n := 1

	var transaction *neoism.Tx
	var err error

	for list.Next(&c) {

		if n%50 == 1 {
			transaction, err = neo.Begin(nil)

			if err != nil {
				panic(err)
			}
		}

		if name, exists := c["name"]; exists {

			props := neoism.Props{}
			sells := false
			cType := "unknown"
			images := []string{}

			for i, value := range c {
				if i == "_id" {
					props["mongodb_id"] = value.(bson.ObjectId).Hex()
					continue
				}

				if i == "store" {
					sells = true
					store := c["store"].(map[string]interface{})
					vendors := store["vendors"].(map[string]interface{})
					spartangeek := vendors["spartangeek"].(map[string]interface{})
					price := spartangeek["price"].(float64)
					stock := spartangeek["stock"].(int)

					props["price"] = price
					props["stock"] = stock

					continue
				}

				if i == "images" {
					images = value.([]string)
					props["images"] = images
					continue
				}

				if i == "type" {
					cType = CamelCase(value.(string))
				}

				props[i] = value
			}

			node, err := neo.CreateNode(props)

			if err != nil {
				panic(err)
			}

			node.AddLabel("Component")
			node.AddLabel(cType)

			if sells {
				node.AddLabel("Product")
			}

			fmt.Println("Processed " + name.(string))
		}

		if n%50 == 0 && transaction != nil {
			err = transaction.Commit()

			if err != nil {
				panic(err)
			}

			transaction = nil
		}

		n++
	}

	if transaction != nil {
		err = transaction.Commit()

		if err != nil {
			panic(err)
		}
	}
}
