package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/amacneil/dbmate/pkg/dbmate"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if len(databaseURL) == 0 {
		log.Fatalln("DATABASE_URL is not set")
	}

	u, err := url.Parse(databaseURL)
	if err != nil {
		log.Fatalln(err)
	}

	db := dbmate.New(u)
	db.MigrationsDir = "./scripts/migrations"

	err = db.Wait()
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Migrate()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Done.")
}
