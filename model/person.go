package model

type Person struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	BirthDate   string `json:"birth_date"`
	BirthPlace  string `json:"birth_place"`
	DeathDate   string `json:"death_date"`
	BurialPlace string `json:"burial_place"`
	FatherID    string `json:"father_id"`
	MotherID    string `json:"mother_id"`
	Bio         string `json:"bio"`
	Note        string `json:"note"`
}