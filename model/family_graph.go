package model

type FamilyGraphNode struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Type     string             `json:"type"` // person / marriage
	Spouses  []SimplePersonNode `json:"spouses,omitempty"`
	Children []FamilyGraphNode  `json:"children,omitempty"`
}

type SimplePersonNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}