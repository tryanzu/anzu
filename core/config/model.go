package config

// Anzu config params struct.
type Anzu struct {
	Site     anzuSite
	Homedir  string
	Security anzuSecurity
	Mail     anzuMail
}

type anzuSecurity struct {
	StrictIPCheck bool `json:"strictIPCheck"`
}

type anzuSite struct {
	Name        string         `json:"name"`
	TitleMotto  string         `json:"titleMotto"`
	Description string         `json:"description"`
	Url         string         `json:"url"`
	LogoUrl     string         `json:"logoUrl"`
	Theme       string         `json:"theme"`
	Nav         []siteLink     `json:"nav"`
	Services    siteServices   `json:"services"`
	Quickstart  siteQuickstart `json:"quickstart"`
}

type siteLink struct {
	Name string `json:"name"`
	Href string `json:"href"`
}

type anzuMail struct {
	Server   string
	User     string
	Password string
	Port     int
	From     string
	ReplyTo  string
}

type siteServices struct {
	Analytics string `json:"-"`
}

type siteQuickstart struct {
	Headline    string           `json:"headline"`
	Description string           `json:"description"`
	Links       []quickstartLink `json:"links"`
}

type quickstartLink struct {
	Name        string `json:"name"`
	Href        string `json:"href"`
	Description string `json:"description"`
}
