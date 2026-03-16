package service

import (
	"database/sql"
	"errors"
	"strings"

	"family-tree/database"
	"family-tree/model"
)

func CreatePerson(p model.Person) error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("person id is required")
	}
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("person name is required")
	}

	query := `
	INSERT INTO people
	(id,name,gender,birth_date,birth_place,death_date,burial_place,father_id,mother_id,bio,note)
	VALUES (?,?,?,?,?,?,?,?,?,?,?)`

	_, err := database.DB.Exec(query,
		p.ID, p.Name, p.Gender, p.BirthDate, p.BirthPlace, p.DeathDate,
		p.BurialPlace, p.FatherID, p.MotherID, p.Bio, p.Note,
	)
	return err
}

func UpdatePerson(p model.Person) error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("person id is required")
	}

	query := `
	UPDATE people SET
		name=?,
		gender=?,
		birth_date=?,
		birth_place=?,
		death_date=?,
		burial_place=?,
		father_id=?,
		mother_id=?,
		bio=?,
		note=?,
		updated_at=CURRENT_TIMESTAMP
	WHERE id=?`

	_, err := database.DB.Exec(query,
		p.Name, p.Gender, p.BirthDate, p.BirthPlace, p.DeathDate,
		p.BurialPlace, p.FatherID, p.MotherID, p.Bio, p.Note, p.ID,
	)
	return err
}

func GetPerson(id string) (*model.Person, error) {
	query := `
	SELECT id,name,gender,birth_date,birth_place,death_date,burial_place,father_id,mother_id,bio,note
	FROM people
	WHERE id=?`

	row := database.DB.QueryRow(query, id)

	var p model.Person
	err := row.Scan(
		&p.ID, &p.Name, &p.Gender, &p.BirthDate, &p.BirthPlace, &p.DeathDate,
		&p.BurialPlace, &p.FatherID, &p.MotherID, &p.Bio, &p.Note,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func SearchPerson(name string) ([]model.Person, error) {
	query := `
	SELECT id,name,gender,birth_date,birth_place,death_date,burial_place,father_id,mother_id,bio,note
	FROM people
	WHERE name LIKE ?
	ORDER BY name,id
	LIMIT 50`

	rows, err := database.DB.Query(query, "%"+name+"%")
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

func GetChildren(id string) ([]model.Person, error) {
	query := `
	SELECT id,name,gender,birth_date,birth_place,death_date,burial_place,father_id,mother_id,bio,note
	FROM people
	WHERE father_id=? OR mother_id=?
	ORDER BY id`

	rows, err := database.DB.Query(query, id, id)
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

func PersonExists(id string) (bool, error) {
	row := database.DB.QueryRow(`SELECT 1 FROM people WHERE id=?`, id)
	var x int
	err := row.Scan(&x)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func toViewPerson(p *model.Person) model.ViewPerson {
	if p == nil {
		return model.ViewPerson{}
	}
	return model.ViewPerson{
		ID:     p.ID,
		Name:   p.Name,
		Gender: p.Gender,
		Note:   p.Note,
	}
}