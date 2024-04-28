package postgres

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

func New(connString string) (*Postgres, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &Postgres{Db: db}, nil
}
