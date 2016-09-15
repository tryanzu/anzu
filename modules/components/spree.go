package components

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
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

const SPREE_SHIPPING_CATEGORY string = "1"
const SPREE_API_URL string = "http://store-1.spartangeek.com/api/v1/"
const SPREE_TOKEN string = "5ac600cefe5a115a775c1adab377214827dfc2e9845ebc11"

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

func (c *ComponentModel) Spree() (*SpreeRecord, error) {

	var rel SpreeRecord
	database := c.di.Mongo.Database
	err := database.C("spree_rels").Find(bson.M{"component_id": c.Id}).One(&rel)

	if err != nil {
		spree_id, err := c.spreeInitialSync()

		if err != nil {
			return nil, err
		}

		rel.Id = bson.NewObjectId()
		rel.SpreeId = spree_id
		rel.ComponentId = c.Id
		rel.Type = c.Type
		rel.Properties = make([]string, 0)
		rel.Images = make([]string, 0)
		rel.Created = time.Now()
		rel.Updated = time.Now()

		err = database.C("spree_rels").Insert(rel)

		if err != nil {
			panic(err)
		}
	}

	point := &rel
	point.SetDI(c.di)

	return point, nil
}

func (c *ComponentModel) spreeInitialSync() (int, error) {

	keywords := []string{c.Manufacturer, c.Name, c.FullName, c.PartNumber}
	form := url.Values{}

	// Set fake price by default
	price := 100000.0

	form.Add("product[name]", c.FullName)
	form.Add("product[meta_keywords]", strings.Join(keywords, ", "))
	form.Add("product[price]", strconv.FormatFloat(price, 'f', 0, 64))
	form.Add("product[shipping_category_id]", SPREE_SHIPPING_CATEGORY)
	form.Add("product[sku]", c.PartNumber)

	req, _ := http.NewRequest("POST", SPREE_API_URL+"products", strings.NewReader(form.Encode()))
	client := http.Client{}

	req.Header.Add("X-Spree-Token", SPREE_TOKEN)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		return 0, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 201 {
		return 0, errors.New(fmt.Sprintf("Could not post product (%v): %v\n\nAborting....\n", resp.StatusCode, string(body)))
	}

	var response struct {
		Id int `json:"id"`
	}

	err = json.Unmarshal(body, &response)

	if err != nil {
		return 0, errors.New(fmt.Sprintf("Could not parse product response: %v\n\nSkipping....\n", err))
	}

	go c.spreeAsyncImages(response.Id)
	go c.spreeAsyncProperties(response.Id)

	return response.Id, nil
}

func (c *ComponentModel) spreeAsyncImages(id int) {
	defer c.di.Errors.Recover()
	database := c.di.Mongo.Database

	if len(c.Images) > 0 {
		for _, image := range c.Images {
			url := "https://assets.spartangeek.com/components/" + image
			imgReader, err := getRemoteImage(url)

			if err != nil {
				panic(err)
			}

			var buff bytes.Buffer
			var fileWr io.Writer

			writter := multipart.NewWriter(&buff)
			writter.SetBoundary("__X_SPARTAN_BOUNDARY__")

			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "image[attachment]", image))
			h.Set("Content-Type", "image/jpg")

			fileWr, err = writter.CreatePart(h)
			if err != nil {
				panic(err)
			}

			if _, err = io.Copy(fileWr, imgReader); err != nil {
				panic(err)
			}

			writter.Close()

			client := http.Client{}
			req, _ := http.NewRequest("POST", SPREE_API_URL+"products/"+strconv.Itoa(id)+"/images", &buff)
			req.Header.Add("X-Spree-Token", SPREE_TOKEN)
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
				panic(fmt.Sprintf("Could not post product image (%v)(%s): %v\n\nAborting....\n", resp.StatusCode, image, string(body)))
			}

			err = database.C("spree_rels").Update(bson.M{"spree_id": id}, bson.M{"$push": bson.M{"images": image}})

			if err != nil {
				panic(err)
			}
		}
	}
}

func (c *ComponentModel) spreeAsyncProperties(id int) {

	defer c.di.Errors.Recover()
	var component map[string]interface{}

	database := c.di.Mongo.Database
	err := database.C("components").Find(bson.M{"_id": c.Id}).One(&component)

	if err != nil {
		panic(err)
	}

	fields, exists := parts[c.Type]
	if !exists {
		return
	}

	for _, field := range fields {
		value, exists := component[field]

		if exists {
			casted := value.(string)
			form := url.Values{}
			form.Add("product_property[property_name]", field)
			form.Add("product_property[value]", casted)

			req, _ := http.NewRequest("POST", SPREE_API_URL+"products/"+strconv.Itoa(id)+"/product_properties", strings.NewReader(form.Encode()))
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

			if resp.StatusCode != 201 {
				panic(fmt.Sprintf("Could not post property (%v): %v\n\nAborting....\n", resp.StatusCode, string(body)))
			}

			err = database.C("spree_rels").Update(bson.M{"spree_id": id}, bson.M{"$push": bson.M{"properties": field}})

			if err != nil {
				panic(err)
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
