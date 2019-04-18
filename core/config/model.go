package config

// Anzu config params struct.
type Anzu struct {
	Site     anzuSite
	Homedir  string
	Security anzuSecurity
	Mail     anzuMail
	Profiler gpcProfiler
}

type gpcProfiler struct {
	Id      string
	Enabled bool
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
	Reactions   [][]string     `json:"reactions"`
}

func (site anzuSite) MakeReactions(names []string) []string {
	m := make(map[string][]string, len(site.Reactions))
	for _, rs := range site.Reactions {
		if len(rs) == 0 {
			continue
		}
		// Assign in map
		m[rs[0]] = rs[1:]
	}
	defaults, hasDefaults := m["default"]
	list := []string{}
	for _, ns := range names {
		if reactions, exists := m[ns]; exists {
			list = append(list, reactions...)
		}
	}
	if len(list) == 0 && hasDefaults {
		list = defaults
	}
	return list
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
