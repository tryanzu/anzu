package feed

type AlgoliaPostModel struct {
	Id         string               `json:"objectID"`
	Title      string               `json:"title"`
	Content    string               `json:"content"`
	Comments   int                  `json:"comments_count"`
	User       AlgoliaUserModel     `json:"user"`
	Tribute    int                  `json:"tribute_count"`
	Shit       int                  `json:"shit_count"`
	Category   AlgoliaCategoryModel `json:"category"`
	Popularity float64              `json:"popularity"`
	Components []string             `json:"components,omitempty"`
	Created    int64                `json:"created"`
}

type AlgoliaCategoryModel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AlgoliaUserModel struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Image    string `json:"image"`
	Email    string `json:"email"`
}
