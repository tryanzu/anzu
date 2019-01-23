package config

type Anzu struct {
	Site anzuSite
	Mail anzuMail
}

type anzuSite struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Url         string       `json:"url"`
	LogoUrl     string       `json:"logoUrl"`
	Nav         []siteLink   `json:"nav"`
	Services    siteServices `json:"services"`
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
}

type siteServices struct {
	Analytics string
}
