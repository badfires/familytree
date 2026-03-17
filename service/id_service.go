package service

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"family-tree/database"
)

const (
	SeqTypePerson   = "person"
	SeqTypeMarriage = "marriage"
	SeqTypeAdoption = "adoption"

	PersonPrefix   = "p"
	MarriagePrefix = "m"
	AdoptionPrefix = "a"
)

var trailingNumberRegexp = regexp.MustCompile(`(\d+)$`)

func EnsureSequencesInitialized() error {
	seqs := []string{
		SeqTypePerson,
		SeqTypeMarriage,
		SeqTypeAdoption,
	}

	for _, seqType := range seqs {
		_, err := database.DB.Exec(`
			INSERT OR IGNORE INTO id_sequences(seq_type, current_value)
			VALUES (?, 0)
		`, seqType)
		if err != nil {
			return err
		}
	}

	return syncSequenceFromExistingData()
}

func syncSequenceFromExistingData() error {
	type seqInit struct {
		seqType string
		table   string
		column  string
	}

	items := []seqInit{
		{SeqTypePerson, "people", "id"},
		{SeqTypeMarriage, "marriages", "id"},
		{SeqTypeAdoption, "adoptions", "id"},
	}

	for _, item := range items {
		maxVal, err := findMaxTrailingNumber(item.table, item.column)
		if err != nil {
			return err
		}

		_, err = database.DB.Exec(`
			UPDATE id_sequences
			SET current_value = CASE
				WHEN current_value < ? THEN ?
				ELSE current_value
			END
			WHERE seq_type = ?
		`, maxVal, maxVal, item.seqType)
		if err != nil {
			return err
		}
	}

	return nil
}

func findMaxTrailingNumber(table, column string) (int, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s`, column, table)
	rows, err := database.DB.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	maxVal := 0
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		n := extractTrailingNumber(id)
		if n > maxVal {
			maxVal = n
		}
	}
	return maxVal, rows.Err()
}

func extractTrailingNumber(id string) int {
	id = strings.TrimSpace(id)
	if id == "" {
		return 0
	}
	matches := trailingNumberRegexp.FindStringSubmatch(id)
	if len(matches) < 2 {
		return 0
	}
	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return n
}

func NextID(tx *sql.Tx, seqType, prefix string) (string, error) {
	_, err := tx.Exec(`
		UPDATE id_sequences
		SET current_value = current_value + 1
		WHERE seq_type = ?
	`, seqType)
	if err != nil {
		return "", err
	}

	var current int
	err = tx.QueryRow(`
		SELECT current_value
		FROM id_sequences
		WHERE seq_type = ?
	`, seqType).Scan(&current)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%d", prefix, current), nil
}