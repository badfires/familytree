package service

import (
	"database/sql"
	"errors"
	"strings"

	"family-tree/database"
	"family-tree/model"
)

func CreateAdoption(a model.Adoption) (*model.Adoption, error) {
	personID := strings.TrimSpace(a.PersonID)
	toFatherID := strings.TrimSpace(a.ToFatherID)
	toMotherID := strings.TrimSpace(a.ToMotherID)

	if personID == "" {
		return nil, errors.New("person id is required")
	}
	if toFatherID == "" && toMotherID == "" {
		return nil, errors.New("to_father_id or to_mother_id is required")
	}

	exists, err := PersonExists(personID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("person does not exist")
	}

	person, err := GetPerson(personID)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, errors.New("person does not exist")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 自动记录原父母
	a.PersonID = personID
	a.FromFatherID = strings.TrimSpace(person.FatherID)
	a.FromMotherID = strings.TrimSpace(person.MotherID)
	a.ToFatherID = toFatherID
	a.ToMotherID = toMotherID

	// 已存在过继记录则更新，否则创建
	existing, err := getAdoptionTx(tx, personID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		a.ID = existing.ID
		_, err = tx.Exec(`
			UPDATE adoptions
			   SET from_father_id = ?,
			       from_mother_id = ?,
			       to_father_id   = ?,
			       to_mother_id   = ?,
			       note           = ?,
			       updated_at     = CURRENT_TIMESTAMP
			 WHERE person_id = ?
		`,
			a.FromFatherID,
			a.FromMotherID,
			a.ToFatherID,
			a.ToMotherID,
			a.Note,
			a.PersonID,
		)
		if err != nil {
			return nil, err
		}
	} else {
		newID, err := NextID(tx, SeqTypeAdoption, AdoptionPrefix)
		if err != nil {
			return nil, err
		}
		a.ID = newID

		_, err = tx.Exec(`
			INSERT INTO adoptions
				(id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note, updated_at)
			VALUES
				(?,?,?,?,?,?,?,CURRENT_TIMESTAMP)
		`,
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
	}

	// 同步把 people 当前父母改成过继后的父母
	_, err = tx.Exec(`
		UPDATE people
		   SET father_id = ?,
		       mother_id = ?
		 WHERE id = ?
	`, a.ToFatherID, a.ToMotherID, a.PersonID)
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

func getAdoptionTx(tx *sql.Tx, personID string) (*model.Adoption, error) {
	row := tx.QueryRow(`
		SELECT id, person_id, from_father_id, from_mother_id, to_father_id, to_mother_id, note
		FROM adoptions
		WHERE person_id = ?
	`, personID)

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