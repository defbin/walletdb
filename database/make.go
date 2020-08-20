package database

import (
	"database/sql"

	"github.com/zeebo/errs"
)

func OpenDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, errs.Combine(db.Close(), err)
	}

	return db, nil
}
