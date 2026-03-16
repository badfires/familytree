package model

type FamilyView struct {
	Center    ViewPerson         `json:"center"`
	Parents   []ViewPerson       `json:"parents"`
	Marriages []ViewMarriageNode `json:"marriages"`
	Adoption  *AdoptionView      `json:"adoption,omitempty"`
}

type ViewPerson struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Note   string `json:"note"`
}

type ViewMarriageNode struct {
	MarriageID string       `json:"marriage_id"`
	Spouses    []ViewPerson `json:"spouses"`
	Children   []ViewPerson `json:"children"`
}

type AdoptionView struct {
	PersonID string       `json:"person_id"`
	From     []ViewPerson `json:"from"`
	To       []ViewPerson `json:"to"`
	Note     string       `json:"note"`
}