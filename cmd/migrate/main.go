package main

import (
	"log"
	"os"

	"stripe-go-spike/internal/db"
)

func main() {
	// Connect based on env (TURSO_DATABASE_URL for Turso, otherwise local SQLite)
	database, err := db.NewConnection()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	mfs := os.DirFS("db/migrations")
	if err := db.RunMigrations(database, mfs, ""); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations applied successfully")
}
