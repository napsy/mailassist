package main

import (
	"fmt"
	"hash/crc32"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func hashMail(date, from, subject string) string {
	str := date + from + subject
	h := crc32.ChecksumIEEE([]byte(str))
	return fmt.Sprintf("%X", h)
}

type localDB struct {
	read []string
}

func newLocalDB() *localDB {
	return &localDB{}
}

func (db *localDB) wasRead(hash string) bool {
	for i := range db.read {
		if db.read[i] == hash {
			return true
		}
	}
	return false
}
func (db *localDB) markRead(hash string) {
	db.read = append(db.read, hash)
}

type sqlMessage struct {
	ID      uint
	Date    time.Time
	From    string
	Subject string

	Summary  string
	Original string

	// Metadata
	Tags    []string
	Deleted bool
}
type sqliteDB struct {
	db          *gorm.DB
	showDeleted bool
}

func newSqlite() (*sqliteDB, error) {
	db, err := gorm.Open(sqlite.Open("mailassist.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	db.AutoMigrate(&sqlMessage{})
	/*
		// Create
		db.Create(&Product{Code: "D42", Price: 100})

		// Read
		var product Product
		db.First(&product, 1)                 // find product with integer primary key
		db.First(&product, "code = ?", "D42") // find product with code D42

		// Update - update product's price to 200
		db.Model(&product).Update("Price", 200)
		// Update - update multiple fields
		db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
		db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

		// Delete - delete product
		db.Delete(&product, 1)
	*/
	return &sqliteDB{
		db: db,
	}, nil
}

func (db *sqliteDB) saveMessage(date, from, subject, message, original string) error {
	// Fri, 29 Mar 2024 17:48:24 +0000 (UTC)
	// Mon Jan 2 15:04:05 MST 2006
	d, err := time.Parse("Mon, 02 Jan 2006 15:04:05 +0000 (MST)", date)
	if err != nil {
		return err
	}
	db.db.Create(&sqlMessage{
		Date:     d,
		From:     from,
		Subject:  subject,
		Original: original,
		Summary:  message,
	})
	return nil
}
func (db *sqliteDB) getMessages(from, to time.Time) []sqlMessage {
	return []sqlMessage{}
}

func (db *sqliteDB) getTags(tags []string) []sqlMessage {
	return []sqlMessage{}
}
