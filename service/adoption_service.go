package service

import (
	"database/sql"
	"errors"
	"strings"

	"family-tree/database"
	"family-tree/model"
)

func CreateAdoption(a model.Adoption) (*model.Adoption, error) {
	if strings.TrimSpace(a.PersonID) == "" {
		return nil, errors.New("person id is required")
	}

	exists, err := PersonExists(a.PersonID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("person does not exist")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	newID, err := NextID(tx, SeqTypeAdoption, AdoptionPrefix)
	if err != nil {
		return nil, err
	}
	a.ID = newID

	query := `
		INSERT OR REPLACE INTO adoptions
			(id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note, updated_at)
		VALUES
			(?,?,?,?,?,?,?,CURRENT_TIMESTAMP)
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

func GetAdoption(personID string) (*model.Adoption, error) {
	query := `
		SELECT id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note
		FROM adoptions
		WHERE person_id = ?
	`
	row := database.DB.QueryRow(query, personID)

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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &a, nil
}

func GetAdoptionsByAdoptiveParent(parentID string) ([]model.Adoption, error) {
	query := `
		SELECT id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note
		FROM adoptions
		WHERE to_father_id = ? OR to_mother_id = ?
		ORDER BY person_id
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

func GetAdoptionsByAdoptivePair(toFatherID, toMotherID string) ([]model.Adoption, error) {
	query := `
		SELECT id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note
		FROM adoptions
		WHERE ifnull(to_father_id, '') = ifnull(?, '')
		  AND ifnull(to_mother_id, '') = ifnull(?, '')
		ORDER BY person_id
	`
	rows, err := database.DB.Query(query, toFatherID, toMotherID)
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