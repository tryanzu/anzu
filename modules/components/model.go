package components

type AlgoliaComponentModel struct {
	Id       string `json:"objectID"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Part     string `json:"part_number"`
	Slug     string `json:"slug"`
	Image    string `json:"image"`
}