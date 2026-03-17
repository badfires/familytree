package service

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"family-tree/database"
	"family-tree/model"
)

var personIDRegexp = regexp.MustCompile(`^p(\d+)$`)

type PersonCSVImportResult struct {
	Success  bool   `json:"success"`
	Imported int    `json:"imported"`
	Message  string `json:"message,omitempty"`
	Row      int    `json:"row,omitempty"`
}

func GetNextPersonID() (string, error) {
	rows, err := database.DB.Query(`SELECT id FROM people`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	maxVal := 0
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		n := extractPersonIDNumber(id)
		if n > maxVal {
			maxVal = n
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return fmt.Sprintf("p%d", maxVal+1), nil
}

func extractPersonIDNumber(id string) int {
	id = strings.TrimSpace(id)
	matches := personIDRegexp.FindStringSubmatch(id)
	if len(matches) != 2 {
		return 0
	}
	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return n
}

func BuildPersonTemplateCSV() ([]byte, error) {
	nextID, err := GetNextPersonID()
	if err != nil {
		return nil, err
	}
	start := extractPersonIDNumber(nextID)
	if start <= 0 {
		start = 1
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	header := []string{
		"id",
		"name",
		"gender",
		"birth_date",
		"birth_place",
		"death_date",
		"burial_place",
		"father_id",
		"mother_id",
		"bio",
		"note",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	sampleRows := [][]string{
		{fmt.Sprintf("p%d", start), "张三", "male", "1950-01-01", "北京", "", "", "", "", "人物简介", "备注"},
		{fmt.Sprintf("p%d", start+1), "李四", "female", "1952-03-01", "上海", "", "", "", "", "", ""},
		{fmt.Sprintf("p%d", start+2), "张小明", "male", "1980-06-01", "广州", "", "", fmt.Sprintf("p%d", start), fmt.Sprintf("p%d", start+1), "", ""},
	}
	for _, row := range sampleRows {
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ImportPeopleCSV(r io.Reader) (*PersonCSVImportResult, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true

	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, errors.New("csv内容为空")
	}

	expectedHeader := []string{
		"id",
		"name",
		"gender",
		"birth_date",
		"birth_place",
		"death_date",
		"burial_place",
		"father_id",
		"mother_id",
		"bio",
		"note",
	}
	if err := validateHeader(rows[0], expectedHeader); err != nil {
		return nil, err
	}

	existingIDs, err := loadExistingPersonIDs()
	if err != nil {
		return nil, err
	}

	type personRow struct {
		rowNum int
		person model.Person
	}
	var parsed []personRow
	csvIDs := make(map[string]struct{})
	maxImportedPersonNo := 0

	for i := 1; i < len(rows); i++ {
		rowNum := i + 1
		row := normalizeRow(rows[i], len(expectedHeader))

		p := model.Person{
			ID:          strings.TrimSpace(row[0]),
			Name:        strings.TrimSpace(row[1]),
			Gender:      strings.TrimSpace(row[2]),
			BirthDate:   strings.TrimSpace(row[3]),
			BirthPlace:  strings.TrimSpace(row[4]),
			DeathDate:   strings.TrimSpace(row[5]),
			BurialPlace: strings.TrimSpace(row[6]),
			FatherID:    strings.TrimSpace(row[7]),
			MotherID:    strings.TrimSpace(row[8]),
			Bio:         strings.TrimSpace(row[9]),
			Note:        strings.TrimSpace(row[10]),
		}

		if isEmptyPersonRow(p) {
			continue
		}
		if p.ID == "" {
			return &PersonCSVImportResult{Success: false, Row: rowNum}, fmt.Errorf("第%d行: id不能为空", rowNum)
		}
		if !personIDRegexp.MatchString(p.ID) {
			return &PersonCSVImportResult{Success: false, Row: rowNum}, fmt.Errorf("第%d行: id格式必须为 p数字，例如 p38", rowNum)
		}
		if p.Name == "" {
			return &PersonCSVImportResult{Success: false, Row: rowNum}, fmt.Errorf("第%d行: name不能为空", rowNum)
		}
		if _, ok := csvIDs[p.ID]; ok {
			return &PersonCSVImportResult{Success: false, Row: rowNum}, fmt.Errorf("第%d行: csv内id重复: %s", rowNum, p.ID)
		}
		if _, ok := existingIDs[p.ID]; ok {
			return &PersonCSVImportResult{Success: false, Row: rowNum}, fmt.Errorf("第%d行: id已存在于系统中: %s", rowNum, p.ID)
		}

		n := extractPersonIDNumber(p.ID)
		if n > maxImportedPersonNo {
			maxImportedPersonNo = n
		}

		csvIDs[p.ID] = struct{}{}
		parsed = append(parsed, personRow{rowNum: rowNum, person: p})
	}

	allKnownIDs := make(map[string]struct{}, len(existingIDs)+len(csvIDs))
	for id := range existingIDs {
		allKnownIDs[id] = struct{}{}
	}
	for id := range csvIDs {
		allKnownIDs[id] = struct{}{}
	}

	for _, item := range parsed {
		p := item.person

		if p.FatherID != "" {
			if _, ok := allKnownIDs[p.FatherID]; !ok {
				return &PersonCSVImportResult{Success: false, Row: item.rowNum}, fmt.Errorf("第%d行: father_id不存在: %s", item.rowNum, p.FatherID)
			}
		}
		if p.MotherID != "" {
			if _, ok := allKnownIDs[p.MotherID]; !ok {
				return &PersonCSVImportResult{Success: false, Row: item.rowNum}, fmt.Errorf("第%d行: mother_id不存在: %s", item.rowNum, p.MotherID)
			}
		}
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	insertPersonSQL := `
		INSERT INTO people (
			id,name,gender,birth_date,birth_place,death_date,burial_place,
			father_id,mother_id,bio,note
		) VALUES (?,?,?,?,?,?,?,?,?,?,?)
	`

	imported := 0
	for _, item := range parsed {
		p := item.person
		_, err := tx.Exec(
			insertPersonSQL,
			p.ID,
			p.Name,
			p.Gender,
			p.BirthDate,
			p.BirthPlace,
			p.DeathDate,
			p.BurialPlace,
			p.FatherID,
			p.MotherID,
			p.Bio,
			p.Note,
		)
		if err != nil {
			return &PersonCSVImportResult{Success: false, Row: item.rowNum}, fmt.Errorf("第%d行插入失败: %w", item.rowNum, err)
		}
		imported++
	}

	// 把 person 序列推进到导入后的最大值，避免后续 CreatePerson 生成重复 ID
	if err := bumpSequenceToAtLeast(tx, SeqTypePerson, maxImportedPersonNo); err != nil {
		return nil, err
	}

	// 自动补 marriage + marriage_children
	for _, item := range parsed {
		p := item.person
		if p.FatherID == "" || p.MotherID == "" {
			continue
		}

		marriageID, err := findMarriageIDByParentsTx(tx, p.FatherID, p.MotherID)
		if err != nil {
			return &PersonCSVImportResult{Success: false, Row: item.rowNum}, fmt.Errorf("第%d行处理婚姻关系失败: %w", item.rowNum, err)
		}

		if marriageID == "" {
			marriageID, err = createMarriageTx(tx, p.FatherID, p.MotherID, "", "auto created by csv import")
			if err != nil {
				return &PersonCSVImportResult{Success: false, Row: item.rowNum}, fmt.Errorf("第%d行自动创建婚姻失败: %w", item.rowNum, err)
			}
		}

		if err := addChildToMarriageTx(tx, marriageID, p.ID); err != nil {
			return &PersonCSVImportResult{Success: false, Row: item.rowNum}, fmt.Errorf("第%d行绑定婚姻子女关系失败: %w", item.rowNum, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &PersonCSVImportResult{
		Success:  true,
		Imported: imported,
		Message:  fmt.Sprintf("成功导入 %d 条人物数据，并自动补齐父母婚姻关系", imported),
	}, nil
}

func validateHeader(actual, expected []string) error {
	if len(actual) < len(expected) {
		return fmt.Errorf("csv表头列数不正确")
	}
	for i := range expected {
		if strings.TrimSpace(actual[i]) != expected[i] {
			return fmt.Errorf("csv表头不正确，第%d列应为 %s，实际为 %s", i+1, expected[i], strings.TrimSpace(actual[i]))
		}
	}
	return nil
}

func normalizeRow(row []string, expectedLen int) []string {
	if len(row) >= expectedLen {
		return row[:expectedLen]
	}
	out := make([]string, expectedLen)
	copy(out, row)
	return out
}

func isEmptyPersonRow(p model.Person) bool {
	return p.ID == "" &&
		p.Name == "" &&
		p.Gender == "" &&
		p.BirthDate == "" &&
		p.BirthPlace == "" &&
		p.DeathDate == "" &&
		p.BurialPlace == "" &&
		p.FatherID == "" &&
		p.MotherID == "" &&
		p.Bio == "" &&
		p.Note == ""
}

func loadExistingPersonIDs() (map[string]struct{}, error) {
	rows, err := database.DB.Query(`SELECT id FROM people`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = struct{}{}
	}
	return result, rows.Err()
}

func bumpSequenceToAtLeast(tx *sql.Tx, seqType string, target int) error {
	if target <= 0 {
		return nil
	}
	_, err := tx.Exec(`
		UPDATE id_sequences
		SET current_value = CASE
			WHEN current_value < ? THEN ?
			ELSE current_value
		END
		WHERE seq_type = ?
	`, target, target, seqType)
	return err
}

func findMarriageIDByParentsTx(tx *sql.Tx, fatherID, motherID string) (string, error) {
	var id string
	err := tx.QueryRow(`
		SELECT id
		FROM marriages
		WHERE (husband_id = ? AND wife_id = ?)
		   OR (husband_id = ? AND wife_id = ?)
		ORDER BY id
		LIMIT 1
	`, fatherID, motherID, motherID, fatherID).Scan(&id)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

func createMarriageTx(tx *sql.Tx, fatherID, motherID, marriageDate, note string) (string, error) {
	newID, err := NextID(tx, SeqTypeMarriage, MarriagePrefix)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(`
		INSERT INTO marriages (id, husband_id, wife_id, marriage_date, note)
		VALUES (?, ?, ?, ?, ?)
	`, newID, fatherID, motherID, marriageDate, note)
	if err != nil {
		return "", err
	}

	return newID, nil
}

func addChildToMarriageTx(tx *sql.Tx, marriageID, childID string) error {
	_, err := tx.Exec(`
		INSERT OR IGNORE INTO marriage_children (marriage_id, child_id)
		VALUES (?, ?)
	`, marriageID, childID)
	return err
}