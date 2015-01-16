package main

import (
    "io/ioutil"
    "errors"
    "strings"
    "net/http"
    "net/url"
    "fmt"
    "crypto/tls"
)

func downloadFromUrl(from string) error {

    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}

    // Parse the filename
    u, err := url.Parse(from)
	tokens := strings.Split(u.Path, "/")
	fileName := tokens[len(tokens)-1]

	response, err := client.Get(from)
	if err != nil {
		return errors.New(fmt.Sprint("Error while downloading", from, "-", err))
	}

	// Read all the bytes to the image
	data, err := ioutil.ReadAll(response.Body)
    if err != nil {
    	return errors.New(fmt.Sprint("Error while downloading", from, "-", err))
    }

	// Close the request channel when needed
	response.Body.Close()

    assets_path := config["paths"].(map[string]interface{})["assets"].(string)
	ioutil.WriteFile(assets_path + fileName, data, 0666)

	return nil
}