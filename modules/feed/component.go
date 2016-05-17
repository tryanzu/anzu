package feed

type Components struct {
	Cpu               Component `bson:"cpu,omitempty" json:"cpu,omitempty"`
	Motherboard       Component `bson:"motherboard,omitempty" json:"motherboard,omitempty"`
	Ram               Component `bson:"ram,omitempty" json:"ram,omitempty"`
	Storage           Component `bson:"storage,omitempty" json:"storage,omitempty"`
	Cooler            Component `bson:"cooler,omitempty" json:"cooler,omitempty"`
	Power             Component `bson:"power,omitempty" json:"power,omitempty"`
	Cabinet           Component `bson:"cabinet,omitempty" json:"cabinet,omitempty"`
	Screen            Component `bson:"screen,omitempty" json:"screen,omitempty"`
	Videocard         Component `bson:"videocard,omitempty" json:"videocard,omitempty"`
	Software          string    `bson:"software,omitempty" json:"software,omitempty"`
	Budget            string    `bson:"budget,omitempty" json:"budget,omitempty"`
	BudgetCurrency    string    `bson:"budget_currency,omitempty" json:"budget_currency,omitempty"`
	BudgetType        string    `bson:"budget_type,omitempty" json:"budget_type,omitempty"`
	BudgetFlexibility string    `bson:"budget_flexibility,omitempty" json:"budget_flexibility,omitempty"`
}

type Component struct {
	Content string `bson:"content" json:"content"`
	//Elections bool             `bson:"elections" json:"elections"`
	//Options   []ElectionOption `bson:"options,omitempty" json:"options"`
	Votes  Votes  `bson:"votes" json:"votes"`
	Status string `bson:"status" json:"status"`
	Voted  string `bson:"voted,omitempty" json:"voted,omitempty"`
}
