package model

type Marriage struct {
	ID           string `json:"id"`
	HusbandID    string `json:"husband_id"`
	WifeID       string `json:"wife_id"`
	MarriageDate string `json:"marriage_date"`
	Note         string `json:"note"`
}