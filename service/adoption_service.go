package service

import (
	"database/sql"
	"errors"
	"strings"

	"family-tree/database"
	"family-tree/model"
)

func normalizeAdoption(a model.Adoption) model.Adoption {
	a.ID = strings.TrimSpace(a.ID)
	a.PersonID = strings.TrimSpace(a.PersonID)
	a.FromFatherID = strings.TrimSpace(a.FromFatherID)
	a.FromMotherID = strings.TrimSpace(a.FromMotherID)
	a.ToFatherID = strings.TrimSpace(a.ToFatherID)
	a.ToMotherID = strings.TrimSpace(a.ToMotherID)
	a.Note = strings.TrimSpace(a.Note)
	return a
}

func fillAdoptionFromCurrentParents(a *model.Adoption) error {
	p, err := GetPerson(a.PersonID)
	if err != nil {
		return err
	}
	if a.FromFatherID == "" {
		a.FromFatherID = strings.TrimSpace(p.FatherID)
	}
	if a.FromMotherID == "" {
		a.FromMotherID = strings.TrimSpace(p.MotherID)
	}
	return nil
}

func CreateAdoption(a model.Adoption) (*model.Adoption, error) {
	a = normalizeAdoption(a)

	if a.PersonID == "" {
		return nil, errors.New("person id is required")
	}
	if a.ToFatherID == "" && a.ToMotherID == "" {
		return nil, errors.New("to_father_id or to_mother_id is required")
	}

	exists, err := PersonExists(a.PersonID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("person does not exist")
	}

	if err := fillAdoptionFromCurrentParents(&a); err != nil {
		return nil, err
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	existed, err := getAdoptionByPersonIDTx(tx, a.PersonID)
	if err != nil {
		return nil, err
	}
	if existed != nil {
		return existed, nil
	}

	newID, err := NextID(tx, SeqTypeAdoption, AdoptionPrefix)
	if err != nil {
		return nil, err
	}
	a.ID = newID

	query := `
INSERT INTO adoptions (
    id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
`
	_, err = tx.Exec(
		query,
		a.ID,
		a.PersonID,
		a.FromFatherID,
		a.FromMotherID,
		a.ToFatherID,
		a.ToMotherID,
		a.Note,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &a, nil
}

func UpdateAdoption(a model.Adoption) error {
	a = normalizeAdoption(a)

	if a.ID == "" {
		return errors.New("adoption id is required")
	}
	if a.PersonID == "" {
		return errors.New("person id is required")
	}
	if a.ToFatherID == "" && a.ToMotherID == "" {
		return errors.New("to_father_id or to_mother_id is required")
	}

	if err := fillAdoptionFromCurrentParents(&a); err != nil {
		return err
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	existed, err := getAdoptionByPersonIDTx(tx, a.PersonID)
	if err != nil {
		return err
	}
	if existed != nil && existed.ID != a.ID {
		return errors.New("another adoption already exists for this person: " + existed.ID)
	}

	query := `
UPDATE adoptions
SET person_id = ?,
    from_father_id = ?,
    from_mother_id = ?,
    to_father_id = ?,
    to_mother_id = ?,
    note = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`
	_, err = tx.Exec(
		query,
		a.PersonID,
		a.FromFatherID,
		a.FromMotherID,
		a.ToFatherID,
		a.ToMotherID,
		a.Note,
		a.ID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetAdoption(personID string) (*model.Adoption, error) {
	query := `
SELECT id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note
FROM adoptions
WHERE person_id = ?
`
	row := database.DB.QueryRow(query, strings.TrimSpace(personID))

	var a model.Adoption
	if err := row.Scan(
		&a.ID,
		&a.PersonID,
		&a.FromFatherID,
		&a.FromMotherID,
		&a.ToFatherID,
		&a.ToMotherID,
		&a.Note,
	); err != nil {
		return nil, err
	}
	return &a, nil
}

func GetAdoptionByID(id string) (*model.Adoption, error) {
	query := `
SELECT id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note
FROM adoptions
WHERE id = ?
`
	row := database.DB.QueryRow(query, strings.TrimSpace(id))

	var a model.Adoption
	if err := row.Scan(
		&a.ID,
		&a.PersonID,
		&a.FromFatherID,
		&a.FromMotherID,
		&a.ToFatherID,
		&a.ToMotherID,
		&a.Note,
	); err != nil {
		return nil, err
	}
	return &a, nil
}

func getAdoptionByPersonIDTx(tx *sql.Tx, personID string) (*model.Adoption, error) {
	query := `
SELECT id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note
FROM adoptions
WHERE person_id = ?
LIMIT 1
`
	row := tx.QueryRow(query, strings.TrimSpace(personID))

	var a model.Adoption
	err := row.Scan(
		&a.ID,
		&a.PersonID,
		&a.FromFatherID,
		&a.FromMotherID,
		&a.ToFatherID,
		&a.ToMotherID,
		&a.Note,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}
func GetAdoptionsByAdoptivePair(fatherID, motherID string) ([]model.Adoption, error) {
	fatherID = strings.TrimSpace(fatherID)
	motherID = strings.TrimSpace(motherID)

	query := `
SELECT id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note
FROM adoptions
WHERE COALESCE(to_father_id, '') = COALESCE(?, '')
  AND COALESCE(to_mother_id, '') = COALESCE(?, '')
ORDER BY id
`
	rows, err := database.DB.Query(query, fatherID, motherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Adoption
	for rows.Next() {
		var a model.Adoption
		if err := rows.Scan(
			&a.ID,
			&a.PersonID,
			&a.FromFatherID,
			&a.FromMotherID,
			&a.ToFatherID,
			&a.ToMotherID,
			&a.Note,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func GetAdoptionsByAdoptiveParent(parentID string) ([]model.Adoption, error) {
	parentID = strings.TrimSpace(parentID)

	query := `
SELECT id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note
FROM adoptions
WHERE to_father_id = ? OR to_mother_id = ?
ORDER BY id
`
	rows, err := database.DB.Query(query, parentID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Adoption
	for rows.Next() {
		var a model.Adoption
		if err := rows.Scan(
			&a.ID,
			&a.PersonID,
			&a.FromFatherID,
			&a.FromMotherID,
			&a.ToFatherID,
			&a.ToMotherID,
			&a.Note,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}