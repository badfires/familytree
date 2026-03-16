package service

import (
	"errors"
	"strings"

	"family-tree/database"
	"family-tree/model"
)

func CreateMarriage(m model.Marriage) error {
	if strings.TrimSpace(m.ID) == "" {
		return errors.New("marriage id is required")
	}

	query := `
	INSERT INTO marriages (id,husband_id,wife_id,marriage_date,note)
	VALUES (?,?,?,?,?)`

	_, err := database.DB.Exec(query,
		m.ID, m.HusbandID, m.WifeID, m.MarriageDate, m.Note,
	)
	return err
}

func GetMarriageByID(id string) (*model.Marriage, error) {
	query := `
	SELECT id,husband_id,wife_id,marriage_date,note
	FROM marriages
	WHERE id=?`

	row := database.DB.QueryRow(query, id)

	var m model.Marriage
	if err := row.Scan(&m.ID, &m.HusbandID, &m.WifeID, &m.MarriageDate, &m.Note); err != nil {
		return nil, err
	}
	return &m, nil
}

func UpdateMarriage(m model.Marriage) error {
	query := `
	UPDATE marriages
	SET husband_id=?, wife_id=?, marriage_date=?, note=?, updated_at=CURRENT_TIMESTAMP
	WHERE id=?`

	_, err := database.DB.Exec(query,
		m.HusbandID, m.WifeID, m.MarriageDate, m.Note, m.ID,
	)
	return err
}

func GetSpouses(personID string) ([]model.Person, error) {
	query := `
	SELECT DISTINCT p.id,p.name,p.gender,p.birth_date,p.birth_place,p.death_date,p.burial_place,p.father_id,p.mother_id,p.bio,p.note
	FROM marriages m
	JOIN people p ON (p.id = m.husband_id OR p.id = m.wife_id)
	WHERE (m.husband_id=? OR m.wife_id=?)
	AND p.id != ?
	ORDER BY p.id`

	rows, err := database.DB.Query(query, personID, personID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Person
	for rows.Next() {
		var p model.Person
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Gender, &p.BirthDate, &p.BirthPlace, &p.DeathDate,
			&p.BurialPlace, &p.FatherID, &p.MotherID, &p.Bio, &p.Note,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func GetMarriagesByPersonID(personID string) ([]model.Marriage, error) {
	query := `
	SELECT id,husband_id,wife_id,marriage_date,note
	FROM marriages
	WHERE husband_id=? OR wife_id=?
	ORDER BY id`

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
	VALUES (?,?)`
	_, err := database.DB.Exec(query, marriageID, childID)
	return err
}

func GetMarriageChildren(marriageID string) ([]model.Person, error) {
	query := `
	SELECT p.id,p.name,p.gender,p.birth_date,p.birth_place,p.death_date,p.burial_place,p.father_id,p.mother_id,p.bio,p.note
	FROM marriage_children mc
	JOIN people p ON p.id = mc.child_id
	WHERE mc.marriage_id=?
	ORDER BY p.id`

	rows, err := database.DB.Query(query, marriageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Person
	for rows.Next() {
		var p model.Person
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Gender, &p.BirthDate, &p.BirthPlace, &p.DeathDate,
			&p.BurialPlace, &p.FatherID, &p.MotherID, &p.Bio, &p.Note,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}