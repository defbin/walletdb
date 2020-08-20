package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if len(databaseURL) == 0 {
		log.Fatalln("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	seed, err := ioutil.ReadFile("./scripts/seed.sql")
	if err != nil {
		log.Fatalln(err)
	}

	_, err = db.Exec(string(seed))
	if err != nil {
		log.Fatalln(err)
	}
}
