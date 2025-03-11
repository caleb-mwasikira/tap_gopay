package database

import (
	"database/sql"
	"log"
	"os"

	"github.com/caleb-mwasikira/banking/utils"
	"github.com/go-sql-driver/mysql"
)

var (
	db *sql.DB
)

func init() {
	var err error

	utils.LoadEnvVariables()

	db, err = connectToDatabase()
	if err != nil {
		log.Fatalf("error connecting to database; %v\n", err)
	}

	log.Println("connected to database successfuly")
}

func connectToDatabase() (*sql.DB, error) {
	cfg := mysql.Config{
		User:   os.Getenv("DB_USER"),
		Passwd: os.Getenv("DB_PASS"),
		DBName: os.Getenv("DB_NAME"),
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
