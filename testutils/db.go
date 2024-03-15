package testutils_test

import (
	"database/sql"
	"fmt"
	"testing"
)

func SetupDB(t *testing.T) *sql.DB {
	const (
		host     = "localhost"
		port     = 5432
		user     = "filmotheka_user"
		password = "filmotheka_pass"
		dbname   = "filmotheka_db"
	)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		t.Fatalf("Unable to connect to database: %v", err)
	}

	return db
}

func BrokenSetupDB(t *testing.T) *sql.DB {
	const (
		host     = "localhost"
		port     = 5432
		user     = "filmotheka_user"
		password = "no_pass"
		dbname   = "filmotheka_db"
	)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		t.Fatalf("Unable to connect to database: %v", err)
	}

	return db
}
