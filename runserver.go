package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/defbin/walletdb/database"
	"github.com/defbin/walletdb/web"
)

// todo: make configurable
const ServerAddr = ":8080"

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if len(databaseURL) == 0 {
		log.Fatalln("walletdb: DATABASE_URL is not set")
	}

	db, err := database.OpenDB(databaseURL)
	if err != nil {
		log.Fatalf("walletdb: %v\n", err)
	}
	defer db.Close()

	handler := withDB(db, web.MakeRouter())

	fmt.Println("walletdb: starting")
	log.Fatalf("walletdb: %v\n", http.ListenAndServe(ServerAddr, handler))
}

func withDB(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "walletdb:db", db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
