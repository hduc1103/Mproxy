package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:123456789@tcp(mysql:3306)/proxy?parseTime=true&loc=Local")
	if err != nil {
		log.Printf("Error connecting to the database: %v", err)
		return nil, err
	}
	return db, nil
}
