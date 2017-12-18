package model

import (
	"encoding/xml"
)

type SitemapSet struct {
	XMLName     xml.Name     `xml:"urlset"`
	XMLNs       string       `xml:"xmlns,attr"`
	XSI         string       `xml:"xmlns:xsi,attr"`
	XSILocation string       `xml:"xsi:schemaLocation,attr"`
	Urls        []SitemapUrl `xml:"url"`
}

type SitemapUrl struct {
	Location string `xml:"loc"`
	Updated  string `xml:"lastmod"`
	Priority string `xml:"priority"`
}
