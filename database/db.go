package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"

	"family-tree/storage"
)

var DB *sql.DB

func InitDB() {
	db, err := sql.Open("sqlite", "family.db")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec(storage.Schema); err != nil {
		log.Fatal(err)
	}

	DB = db
}