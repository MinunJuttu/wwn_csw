package character

type Sheet struct {
	Attributes Attributes `json:"attributes"`
	HP         HitPoints  `json:"hp"`
}

type Attributes struct {
	Strength     string `json:"strength"`
	Dexterity    string `json:"dexterity"`
	Constitution string `json:"constitution"`
	Intelligence string `json:"intelligence"`
	Wisdom       string `json:"wisdom"`
	Charisma     string `json:"charisma"`
}

type HitPoints struct {
	Current string `json:"current"`
	Max     string `json:"max"`
}
