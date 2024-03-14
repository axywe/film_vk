package storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host     = "db"
	port     = 5432
	user     = "filmotheka_user"
	password = "filmotheka_pass"
	dbname   = "filmotheka_db"
)

func InitDB() (*sql.DB, error) {
	psqlInfo := "host=" + host + " port=" + fmt.Sprint(port) + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}
