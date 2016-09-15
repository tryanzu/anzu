package cli

import (
	"github.com/fernandez14/spartangeek-blacker/modules/components"
	"github.com/fernandez14/spartangeek-blacker/modules/helpers"
	"gopkg.in/mgo.v2/bson"

	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var parts map[string][]string = map[string][]string{
	"cpu-cooler":   []string{"supported_sockets", "liquid_cooled", "bearing_type", "radiator_size", "noise_level", "fan_rpm"},
	"cpu":          []string{"data_width", "socket", "speed", "cores", "l1_cache", "l2_cache", "l3_cache", "lithography", "thermal_design_power", "thermal_design_power", "hyper_threading", "integrated_graphics"},
	"motherboard":  []string{"form_factor", "cpu", "cpu_socket", "chipset", "memory_slots", "memory_type", "max_memory", "raid_support", "onboard_video", "crossfire_support", "sli_support", "pata_100", "pata_133", "esata_3_gbs", "sata_15_gbs", "sata_3_gbs", "sata_6_gbs", "sata_express", "mini_pci_express_msata", "msata", "onboard_ethernet", "onboard_usb_3"},
	"memory":       []string{"memory_type", "speed", "size", "price_per_gb", "cas", "voltage", "heat_spreader", "ecc", "registered", "color"},
	"storage":      []string{"capacity", "interface", "cache", "rpm", "form_factor", "ssd_controller", "nand_flash_type", "price_per_gb", "gb_per_dollar", "power_loss_protection"},
	"video-card":   []string{"interface", "chipset", "memory_size", "memory_type", "core_clock", "tdp", "fan", "sli_support", "crossfire_support", "length", "support_g_sync", "support_freesync", "dvi_d_dual_link", "hdmi", "vga", "displayport", "mini_displayport"},
	"case":         []string{"case_type", "color", "includes_psu", "internal_25_bays", "internal_35_bays", "internal_525_bays", "motherboard_compatibility", "front_usb_3", "maximum_gpu_length", "dimensions"},
	"power-supply": []string{"psu_type", "wattage", "fans", "modular", "efficiency_certification", "efficiency", "output", "pcie_62_pin_connectors", "pcie_6_pin_connectors", "pcie_8_pin_connectors"},
	"monitor":      []string{"screen_size", "recommended_resolution", "widescreen", "aspect_ratio", "viewing_angle", "display_colors", "brightness", "constrast_ratio", "response_time", "refresh_rate", "ips", "led", "builtin_speakers", "dvi_d_dual_link", "dvi_d_single_link", "dvi_i_dual_link", "displayport", "hdmi", "dvi", "dvi_a", "vga", "component", "s_video", "bnc", "mini_displayport", "supports_freesync", "supports_g_sync"},
	"keyboard":     []string{"design_type", "design_style", "normal_keys", "features", "color", "keyboard_type", "backlit", "mechanical", "switch_type", "includes_mouse", "mouse_hand_orientation", "mouse_color", "mouse_tracking_method"},
	"mouse":        []string{"connection", "tracking_method", "maximum_dpi", "color", "hand_orientation"},
	"headphones":   []string{"headphone_type", "enclosure_type", "color", "channels", "microphone", "impedance", "active_noise_cancelling", "sensitivity", "frequency_response", "connection", "features", "cord_length", "weight"},
}

const SPREE_SHIPPING_CATEGORY string = "1"
const SPREE_API_URL string = "http://store-1.spartangeek.com/api/v1/"

type Track struct {
	Id         bson.ObjectId `bson:"_id,omitempty" json:"id"`
	SpreeId    int           `bson:"spree_id"`
	SpartanId  bson.ObjectId `bson:"spartan_id"`
	Type       string        `bson:"type"`
	Images     []string      `bson:"images"`
	Properties []string      `bson:"properties"`
	Created    time.Time     `bson:"created_at"`
	Updated    time.Time     `bson:"updated_at"`
}

func (module Module) SpreeProducts() {

	var c *components.ComponentModel

	database := module.Mongo.Database
	list := database.C("components").Find(bson.M{"store": bson.M{"$exists": true}}).Iter()

	for list.Next(&c) {

		price, err := c.GetVendorPrice("spartangeek")

		if err != nil {
			fmt.Printf("Could not get price for component: %v\n\nSkipping....\n", err)
			continue
		}

		keywords := []string{c.Manufacturer, c.Name, c.FullName, c.PartNumber}
		form := url.Values{}
		form.Add("product[name]", c.FullName)
		form.Add("product[meta_keywords]", strings.Join(keywords, ", "))
		form.Add("product[price]", strconv.FormatFloat(price, 'f', 0, 64))
		form.Add("product[shipping_category_id]", SPREE_SHIPPING_CATEGORY)

		req, _ := http.NewRequest("POST", SPREE_API_URL+"products", strings.NewReader(form.Encode()))
		client := http.Client{}

		req.Header.Add("X-Spree-Token", "5ac600cefe5a115a775c1adab377214827dfc2e9845ebc11")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)

		if err != nil {
			panic(err)
		}

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			panic(err)
		}

		if resp.StatusCode != 201 {
			fmt.Printf("Could not post product (%v): %v\n\nAborting....\n", resp.StatusCode, string(body))
			break
		}

		var response struct {
			Id int `json:"id"`
		}

		err = json.Unmarshal(body, &response)

		if err != nil {
			fmt.Printf("Could not parse product response: %v\n\nSkipping....\n", err)
			continue
		}

		var track Track

		track.SpreeId = response.Id
		track.SpartanId = c.Id
		track.Type = c.Type
		track.Created = time.Now()
		track.Updated = time.Now()

		err = database.C("spree_products").Insert(track)

		if err != nil {
			fmt.Printf("Could not save product rel: %v\n\nAborting....\n", err)
			break
		}

		fmt.Printf("Saved %s to spree %v\n", c.FullName, track.SpreeId)
	}
}

func (module Module) SpreeProductsImages() {

	var c *components.ComponentModel

	database := module.Mongo.Database
	list := database.C("components").Find(bson.M{"store": bson.M{"$exists": true}}).Iter()

	for list.Next(&c) {

		var track Track
		err := database.C("spree_products").Find(bson.M{"spartan_id": c.Id, "images": bson.M{"$exists": false}}).One(&track)

		if err != nil {
			fmt.Printf("Skipping %v....\n", c.Id.Hex())
			continue
		}

		if len(c.Images) > 0 {
			for _, image := range c.Images {
				if exists, _ := helpers.InArray(image, track.Images); exists {
					continue
				}

				url := "https://assets.spartangeek.com/components/" + image
				imgReader, err := getRemoteImage(url)

				if err != nil {
					fmt.Printf("Could not get image: %v\n\nAborting....\n", err)
					break
				}

				var buff bytes.Buffer
				var fileWr io.Writer

				writter := multipart.NewWriter(&buff)
				writter.SetBoundary("__X_SPARTAN_BOUNDARY__")

				h := make(textproto.MIMEHeader)
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "image[attachment]", image))
				h.Set("Content-Type", "application/octet-stream")

				fileWr, err = writter.CreatePart(h)
				if err != nil {
					panic(err)
				}

				if _, err = io.Copy(fileWr, imgReader); err != nil {
					panic(err)
				}

				writter.Close()

				client := http.Client{}
				req, _ := http.NewRequest("POST", SPREE_API_URL+"products/"+strconv.Itoa(track.SpreeId)+"/images", &buff)
				req.Header.Add("X-Spree-Token", "5ac600cefe5a115a775c1adab377214827dfc2e9845ebc11")
				req.Header.Add("Content-Type", writter.FormDataContentType())

				resp, err := client.Do(req)

				if err != nil {
					panic(err)
				}

				body, err := ioutil.ReadAll(resp.Body)

				if err != nil {
					panic(err)
				}

				if resp.StatusCode != 201 {
					fmt.Printf("Could not post product image (%v)(%s): %v\n\nAborting....\n", resp.StatusCode, image, string(body))
					break
				}

				err = database.C("spree_products").Update(bson.M{"_id": track.Id}, bson.M{"$push": bson.M{"images": image}})

				if err != nil {
					panic(err)
				}

				fmt.Printf("Saved %v to spree %v\n", image, track.SpreeId)
			}
		}
	}
}

func (module Module) SpreeProductsProperties() {

	var c map[string]interface{}

	database := module.Mongo.Database
	list := database.C("components").Find(bson.M{"store": bson.M{"$exists": true}}).Iter()

	for list.Next(&c) {

		var track Track

		id, exists := c["_id"].(bson.ObjectId)
		if !exists {
			fmt.Printf("Could not get id (%v): %v\n\nAborting....\n", id, c)
			break
		}

		err := database.C("spree_products").Find(bson.M{"spartan_id": id, "properties": bson.M{"$exists": false}}).One(&track)
		if err != nil {
			fmt.Printf("Skipping %v....\n", id.Hex())
			continue
		}

		cType, exists := c["type"].(string)
		if !exists {
			fmt.Printf("No type, skip %v....\n", id.Hex())
			continue
		}

		fields, exists := parts[cType]
		if !exists {
			fmt.Printf("No fields for %s, skip %v....\n", cType, id.Hex())
			continue
		}

		fmt.Printf("\n%s (%s): ", id.Hex(), cType)

		for _, field := range fields {
			value, exists := c[field]

			if exists {
				casted := value.(string)
				fmt.Printf("%s..", field)

				form := url.Values{}
				form.Add("product_property[property_name]", field)
				form.Add("product_property[value]", casted)

				req, _ := http.NewRequest("POST", SPREE_API_URL+"products/"+strconv.Itoa(track.SpreeId)+"/product_properties", strings.NewReader(form.Encode()))
				client := http.Client{}

				req.Header.Add("X-Spree-Token", "5ac600cefe5a115a775c1adab377214827dfc2e9845ebc11")
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

				resp, err := client.Do(req)

				if err != nil {
					panic(err)
				}

				body, err := ioutil.ReadAll(resp.Body)

				if err != nil {
					panic(err)
				}

				if resp.StatusCode != 201 {
					fmt.Printf("Could not post property (%v): %v\n\nAborting....\n", resp.StatusCode, string(body))
					break
				}

				err = database.C("spree_products").Update(bson.M{"_id": track.Id}, bson.M{"$push": bson.M{"properties": field}})

				if err != nil {
					panic(err)
				}

				fmt.Printf("ok, ")
			}
		}
	}
}

func getRemoteImage(remoteUrl string) (io.ReadCloser, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Download the file
	response, err := client.Get(remoteUrl)

	if err != nil {
		panic(err)
	}

	return response.Body, nil
}

func (module Module) SpreeProductsFlush() {

	var c Track

	database := module.Mongo.Database
	list := database.C("spree_products").Find(nil).Iter()

	for list.Next(&c) {

		req, _ := http.NewRequest("DELETE", SPREE_API_URL+"products/"+strconv.Itoa(c.SpreeId), nil)
		client := http.Client{}

		req.Header.Add("X-Spree-Token", "5ac600cefe5a115a775c1adab377214827dfc2e9845ebc11")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)

		if err != nil {
			panic(err)
		}

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			panic(err)
		}

		if resp.StatusCode != 204 {
			fmt.Printf("Could not flush product (%v, %v, %v): %v\n\nAborting....\n", resp.StatusCode, c.SpreeId, c.SpartanId.Hex(), string(body))
			break
		}

		err = database.C("spree_products").RemoveId(c.Id)

		if err != nil {
			fmt.Printf("Could not flush product rel: %v\n\nAborting....\n", err)
			break
		}

		fmt.Printf(".")
	}
}
