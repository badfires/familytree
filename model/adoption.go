package model

type Adoption struct {
	ID       string `json:"id"`
	PersonID string `json:"person_id"`

	FromFatherID string `json:"from_father_id"`
	FromMotherID string `json:"from_mother_id"`

	ToFatherID string `json:"to_father_id"`
	ToMotherID string `json:"to_mother_id"`

	Note string `json:"note"`
}