package db

import (
	"database/sql"
	"fmt"
	"kanban-api/config"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

func ConnectDB(config config.Config)  error {
	conString := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", 
	config.Database.Host, config.Database.Port, config.Database.DBName, config.Database.User, config.Database.Password)

	var err error
	db, err = sql.Open("postgres", conString) 
	if err != nil {
		return  err
	} 
	return nil
}

func GetDB() *sql.DB{
	return db
}

func CloseDB() error {
	err := db.Close()
	if err != nil {
		log.Printf("Can't close connection DB: %v", err)
	}
	return nil
}

func DBPing() error {
	err := db.Ping()
	if err != nil {
		return err
	}
	return nil
}