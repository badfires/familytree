package service

import (
	"family-tree/database"
	"family-tree/model"
)

func CreateAdoption(a model.Adoption) error {
	query := `
	INSERT OR REPLACE INTO adoptions
	(id,person_id,from_father_id,from_mother_id,to_father_id,to_mother_id,note,updated_at)
	VALUES (?,?,?,?,?,?,?,CURRENT_TIMESTAMP)`

	_, err := database.DB.Exec(query,
		a.ID, a.PersonID, a.FromFatherID, a.FromMotherID, a.ToFatherID, a.ToMotherID, a.Note,
	)
	return err
}

func GetAdoption(personID string) (*model.Adoption, error) {
	query := `
	SELECT id,person_id,from_father_id,from_mother_id,to_father_id,to_mother_id,note
	FROM adoptions
	WHERE person_id=?`

	row := database.DB.QueryRow(query, personID)

	var a model.Adoption
	if err := row.Scan(
		&a.ID, &a.PersonID, &a.FromFatherID, &a.FromMotherID,
		&a.ToFatherID, &a.ToMotherID, &a.Note,
	); err != nil {
		return nil, err
	}

	return &a, nil
}