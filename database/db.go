package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"

	"family-tree/storage"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

var DB *sql.DB
const key = "leimc5" //简单加密,对db文件做一层简单的加密
func InitDB() {

	escapedKey := url.QueryEscape(key)
	dsn := fmt.Sprintf("family.db?_pragma_key=%s&_pragma_cipher_page_size=4096", escapedKey)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// 强制实际打开，并在 key 不正确时尽早失败
	if _, err := db.Exec("SELECT count(*) FROM sqlite_master;"); err != nil {
		_ = db.Close()
		log.Fatalf("open encrypted db failed: %v", err)
	}

	if _, err := db.Exec(storage.Schema); err != nil {
		_ = db.Close()
		log.Fatal(err)
	}

	DB = db
}