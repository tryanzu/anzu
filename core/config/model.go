package config

import "strings"

// Anzu config params struct.
type Anzu struct {
	Site     anzuSite
	Homedir  string
	Security anzuSecurity
	Mail     anzuMail
	Oauth    OauthConfig
}

type OauthConfig struct {
	Facebook OauthKeys
}

type OauthKeys struct {
	Key      string
	Secret   string
	Callback string
}

type chatChan struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Youtube     string `json:"youtubeVideo"`
	Twitch      string `json:"twitchVideo"`
}

type anzuSecurity struct {
	Secret        string `json:"secret"`
	StrictIPCheck bool   `json:"strictIPCheck"`
}

type Flag struct {
	ShouldRemove string `hcl:"shouldRemove"`
	ShouldBan    string `hcl:"shouldBan"`
}

type Rules struct {
	Reactions  map[string]*ReactionEffect `hcl:"reaction"`
	BanReasons map[string]*BanReason      `hcl:"banReason"`
	Flags      map[string]*Flag           `hcl:"flag"`
}

type anzuSite struct {
	Name           string         `json:"name"`
	TitleMotto     string         `json:"titleMotto"`
	Description    string         `json:"description"`
	Url            string         `json:"url"`
	LogoUrl        string         `json:"logoUrl"`
	Theme          string         `json:"theme"`
	Nav            []siteLink     `json:"nav"`
	Chat           []chatChan     `json:"chat"`
	Services       siteServices   `json:"services"`
	Quickstart     siteQuickstart `json:"quickstart"`
	Reactions      [][]string     `json:"reactions"`
	ThirdPartyAuth []string       `json:"thirdPartyAuth"`
}

func (site anzuSite) MakeURL(url string) string {
	u := site.Url
	if strings.HasSuffix(u, "/") == false {
		u = u + "/"
	}
	return u + url
}

func (site anzuSite) IsValidReaction(name string) bool {
	for _, rs := range site.Reactions {
		if len(rs) == 0 {
			continue
		}
		for _, v := range rs[1:] {
			if v == name {
				return true
			}
		}
	}
	return false
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
