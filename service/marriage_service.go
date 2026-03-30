package service

import (
	"database/sql"
	"errors"
	"strings"

	"family-tree/database"
	"family-tree/model"
)

func normalizeMarriage(m model.Marriage) model.Marriage {
	m.ID = strings.TrimSpace(m.ID)
	m.HusbandID = strings.TrimSpace(m.HusbandID)
	m.WifeID = strings.TrimSpace(m.WifeID)
	m.MarriageDate = strings.TrimSpace(m.MarriageDate)
	m.Note = strings.TrimSpace(m.Note)
	return m
}

func findDuplicateMarriage(tx *sql.Tx, m model.Marriage, excludeID string) (*model.Marriage, error) {
	query := `
		SELECT id, husband_id, wife_id, marriage_date, note
		FROM marriages
		WHERE id <> ?
		  AND (
		        (husband_id = ? AND wife_id = ?)
		     OR (husband_id = ? AND wife_id = ?)
		  )
		  AND COALESCE(marriage_date, '') = COALESCE(?, '')
		LIMIT 1
	`
	row := tx.QueryRow(
		query,
		excludeID,
		m.HusbandID, m.WifeID,
		m.WifeID, m.HusbandID,
		m.MarriageDate,
	)

	var existed model.Marriage
	err := row.Scan(&existed.ID, &existed.HusbandID, &existed.WifeID, &existed.MarriageDate, &existed.Note)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &existed, nil
}

func CreateMarriage(m model.Marriage) (*model.Marriage, error) {
	m = normalizeMarriage(m)

	if m.HusbandID == "" && m.WifeID == "" {
		return nil, errors.New("husband_id or wife_id is required")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	dup, err := findDuplicateMarriage(tx, m, "")
	if err != nil {
		return nil, err
	}
	if dup != nil {
		return dup, nil
	}

	newID, err := NextID(tx, SeqTypeMarriage, MarriagePrefix)
	if err != nil {
		return nil, err
	}
	m.ID = newID

	query := `
		INSERT INTO marriages (id, husband_id, wife_id, marriage_date, note)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(query, m.ID, m.HusbandID, m.WifeID, m.MarriageDate, m.Note)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &m, nil
}

func UpdateMarriage(m model.Marriage) error {
	m = normalizeMarriage(m)

	if m.ID == "" {
		return errors.New("marriage id is required")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	dup, err := findDuplicateMarriage(tx, m, m.ID)
	if err != nil {
		return err
	}
	if dup != nil {
		return errors.New("duplicate marriage already exists: " + dup.ID)
	}

	query := `
		UPDATE marriages
		SET husband_id = ?, wife_id = ?, marriage_date = ?, note = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = tx.Exec(query, m.HusbandID, m.WifeID, m.MarriageDate, m.Note, m.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetMarriageByID(id string) (*model.Marriage, error) {
	query := `
		SELECT id, husband_id, wife_id, marriage_date, note
		FROM marriages
		WHERE id = ?
	`
	row := database.DB.QueryRow(query, strings.TrimSpace(id))

	var m model.Marriage
	if err := row.Scan(&m.ID, &m.HusbandID, &m.WifeID, &m.MarriageDate, &m.Note); err != nil {
		return nil, err
	}
	return &m, nil
}

func GetSpouses(personID string) ([]model.Person, error) {
	query := `
		SELECT DISTINCT
			p.id, p.name, p.gender, p.birth_date, p.birth_place, p.death_date, p.burial_place,
			p.father_id, p.mother_id, p.bio, p.note
		FROM marriages m
		JOIN people p ON (p.id = m.husband_id OR p.id = m.wife_id)
		WHERE (m.husband_id = ? OR m.wife_id = ?)
		  AND p.id != ?
		ORDER BY p.id
	`
	rows, err := database.DB.Query(query, personID, personID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Person
	for rows.Next() {
		var p model.Person
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Gender,
			&p.BirthDate,
			&p.BirthPlace,
			&p.DeathDate,
			&p.BurialPlace,
			&p.FatherID,
			&p.MotherID,
			&p.Bio,
			&p.Note,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func GetMarriagesByPersonID(personID string) ([]model.Marriage, error) {
	query := `
		SELECT id, husband_id, wife_id, marriage_date, note
		FROM marriages
		WHERE husband_id = ? OR wife_id = ?
		ORDER BY id
	`
	rows, err := database.DB.Query(query, personID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Marriage
	for rows.Next() {
		var m model.Marriage
		if err := rows.Scan(&m.ID, &m.HusbandID, &m.WifeID, &m.MarriageDate, &m.Note); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

func AddChildToMarriage(marriageID, childID string) error {
	query := `
		INSERT OR IGNORE INTO marriage_children (marriage_id, child_id)
		VALUES (?, ?)
	`
	_, err := database.DB.Exec(query, marriageID, childID)
	return err
}

func GetMarriageChildren(marriageID string) ([]model.Person, error) {
	query := `
		SELECT
			p.id, p.name, p.gender, p.birth_date, p.birth_place, p.death_date, p.burial_place,
			p.father_id, p.mother_id, p.bio, p.note
		FROM marriage_children mc
		JOIN people p ON p.id = mc.child_id
		WHERE mc.marriage_id = ?
		ORDER BY CAST(SUBSTR(p.id, 2) AS INTEGER) DESC, p.id DESC
	`
	rows, err := database.DB.Query(query, marriageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Person
	for rows.Next() {
		var p model.Person
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Gender,
			&p.BirthDate,
			&p.BirthPlace,
			&p.DeathDate,
			&p.BurialPlace,
			&p.FatherID,
			&p.MotherID,
			&p.Bio,
			&p.Note,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}